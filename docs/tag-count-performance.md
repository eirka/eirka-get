# Tag-count performance: design & roadmap

## Context

The tag-listing endpoint `GET /tags/:ib/:page` (the popularity-sorted "tag cloud") degrades
badly as the `tagmap` table grows. On every cache miss it aggregates the **entire** `tagmap`
table — joining `tagmap → images → posts → threads` and `GROUP BY tag_id` — just to return one
page of ~10 tags. Because the results are `ORDER BY count DESC`, every tag's count has to be
computed before page 1 can be returned, so pagination saves transfer but not work. The same
per-tag count pattern recurs in `models/tagsearch.go` (a correlated subquery per matched tag)
and `models/tag.go` (single-tag header count).

Original query (`models/tags.go`):

```sql
SELECT IFNULL(tag_counts.count, 0) AS count, t.tag_id, t.tag_name, t.tagtype_id
FROM tags t
LEFT JOIN (
  SELECT tm.tag_id, COUNT(DISTINCT tm.image_id) as count
  FROM tagmap tm
  INNER JOIN images i ON tm.image_id = i.image_id
  INNER JOIN posts p ON i.post_id = p.post_id AND p.post_deleted != 1
  INNER JOIN threads th ON p.thread_id = th.thread_id AND th.thread_deleted != 1
  GROUP BY tm.tag_id
) tag_counts ON t.tag_id = tag_counts.tag_id
WHERE t.ib_id = ?
ORDER BY count DESC, t.tag_id ASC
LIMIT ?, ?
```

### Decisions that shape the design

Agreed with the product owner:

- **Count semantics:** a *raw `tagmap` association count* is acceptable for the cloud — it no
  longer needs to exclude soft-deleted posts/threads. **This is the key lever.**
- **Staleness:** acceptable in principle (though the chosen design ends up exact).
- **Scope:** cross-repo changes to the write side (eirka-post / eirka-admin) are acceptable.

### Why raw-count is the unlock

The expensive 4-table join exists *only* to honor the "non-deleted images" filter. Dropping
that requirement collapses the count to `COUNT(*) FROM tagmap WHERE tag_id = ?`, which can be
stored as a denormalized column on `tags` and kept **exactly** correct with a single `±1` per
`tagmap` insert/delete. The painful part of any denormalized counter — post/thread soft-delete
*toggles* (in eirka-admin) that change visibility for many tags in one statement — disappears
entirely, because the count no longer depends on `post_deleted` / `thread_deleted`. The read
then becomes a single-table, index-ordered scan: O(page size), independent of `tagmap` size.

---

## Phase 0 — Stopgap (DONE, eirka-get only)

A same-day, zero-migration improvement that keeps exact semantics, shipped in `models/tags.go`:

- **Pushed `ib_id` into the count subquery** by joining `tags` inside it (`tagmap` has no
  `ib_id`), so the aggregation covers only this board's rows instead of the whole table.
- **`COUNT(*)` instead of `COUNT(DISTINCT tm.image_id)`** — the `tagmap` PK `(image_id, tag_id)`
  plus the 1:1 `image → post → thread` chain guarantee one joined row per tagmap row, so the
  distinct dedup is redundant.

Current query:

```sql
SELECT IFNULL(tag_counts.count, 0) AS count, t.tag_id, t.tag_name, t.tagtype_id
FROM tags t
LEFT JOIN (
  SELECT tm.tag_id, COUNT(*) as count
  FROM tagmap tm
  INNER JOIN tags t2 ON tm.tag_id = t2.tag_id AND t2.ib_id = ?
  INNER JOIN images i ON tm.image_id = i.image_id
  INNER JOIN posts p ON i.post_id = p.post_id AND p.post_deleted != 1
  INNER JOIN threads th ON p.thread_id = th.thread_id AND th.thread_deleted != 1
  GROUP BY tm.tag_id
) tag_counts ON t.tag_id = tag_counts.tag_id
WHERE t.ib_id = ?
ORDER BY count DESC, t.tag_id ASC
LIMIT ?, ?
```

This only reduces the **constant factor** — it is still O(all of this board's tags) per cache
miss because of the popularity sort. It is a bridge, not the fix. (Test `WithArgs` updated for
the extra `ib_id` parameter; `go vet` + `go test ./...` pass.)

---

## Phase 1 — Denormalized `tags.tag_count` (the real fix)

### 1. Schema change

There is no migration framework: the canonical schema is the single bootstrap script
`eirka-post/eirka.sql`, applied via Ansible (`eirka-provision/post.yml`). So changes land in
two places — the canonical DDL and a one-off script for live databases.

- Add to `CREATE TABLE tags` in `eirka-post/eirka.sql`:
  `tag_count int unsigned NOT NULL DEFAULT '0'`
- Add a descending composite index (MySQL 8 supports descending indexes) so the read is fully
  index-ordered: `KEY idx_tags_ibid_count (ib_id, tag_count DESC, tag_id)`
- One-off migration/backfill (new file, e.g. `eirka-post/migrations/tag_count.sql`):

```sql
ALTER TABLE tags
  ADD COLUMN tag_count int unsigned NOT NULL DEFAULT 0,
  ADD KEY idx_tags_ibid_count (ib_id, tag_count DESC, tag_id);

UPDATE tags t
LEFT JOIN (SELECT tag_id, COUNT(*) AS c FROM tagmap GROUP BY tag_id) m
  ON t.tag_id = m.tag_id
SET t.tag_count = IFNULL(m.c, 0);
```

### 2. Maintain the counter at the two tagmap write sites

A search confirmed `tagmap` is written in exactly two places, and there are **no hard deletes
of images/posts/threads** (all soft) — so no surprise cascade decrements. (Re-grep for
`tagmap` across the repos at implementation time to confirm no new write site exists.)

- **Insert** — `eirka-post/models/addtag.go` (`AddTagModel.Post`, the `INSERT into tagmap ...`).
  Already inside a transaction; add before commit:
  `UPDATE tags SET tag_count = tag_count + 1 WHERE tag_id = ?`
  Atomic with the insert — a duplicate `(image_id, tag_id)` rolls the tx back and leaves the
  counter untouched.
- **Delete** — `eirka-admin/models/deleteimagetag.go` (`DeleteImageTagModel.Delete`, the
  `DELETE tm FROM tagmap ...`). Currently a plain handle; wrap delete + decrement in a
  transaction (`db.GetTransaction()`), then:
  `UPDATE tags SET tag_count = GREATEST(tag_count - 1, 0) WHERE tag_id = ?`
- **Tag hard-delete** — `eirka-admin/models/deletetag.go` (`DELETE FROM tags WHERE tag_id = ?`)
  cascade-deletes that tag's `tagmap` rows, but the counter lives on the tag row being removed,
  so **no maintenance is needed**. (There is only one affected tag — the one being deleted.)

### 3. Swap the read query (`eirka-get/models/tags.go`)

Replace the derived-table join with a single-table read:

```sql
SELECT tag_count AS count, tag_id, tag_name, tagtype_id
FROM tags
WHERE ib_id = ?
ORDER BY tag_count DESC, tag_id ASC
LIMIT ?, ?
```

Parameters become `i.Ib, paged.Limit, paged.PerPage`; the scan binding
(`count, tag_id, tag_name, tagtype_id`) is unchanged. The pagination `COUNT(*) FROM tags`
(`models/tags.go:59`) is already cheap and stays.

### 4. Align the other tag-count readers (consistency)

Now that "the number next to a tag" means the raw association count, the other listings should
report the same number:

- `models/tagsearch.go` — replace the correlated `(SELECT COUNT(tagmap.image_id) ...)`
  subquery with `tags.tag_count`. This also removes the per-result subquery, fixing tag
  search's own scaling problem.
- `models/tag.go` — the single-tag **header** count. Switching it to `tag_count` keeps the
  header consistent with the cloud. **Visible trade-off:** this header sits above a list of
  *non-deleted* images, so a raw count can exceed the number of images actually shown. This is
  the divergence the product owner accepted; revisit if the per-tag page should keep filtered
  semantics. The image *list* query in `tag.go` is unchanged regardless.

### 5. Redis cache — unchanged

Existing invalidation of the `tags:<ib>` key (addtag, newtag, deleteimagetag, deletetag,
deletethread, deletepost, updatetag) still works and is worth keeping; misses are now cheap.

---

## Rollout order (matters)

1. Apply schema migration + backfill (column defaults to 0; backfill sets real values).
2. Deploy write-side counter maintenance (eirka-post, eirka-admin) so the column stays correct.
3. Deploy the eirka-get read swap (Phase 1 step 3–4).
4. *(Optional)* schedule a periodic reconciliation as a drift backstop, reusing the backfill
   `UPDATE ... JOIN` — cheap insurance against any future cascade/hard-delete path.

Deploying the read swap before write-side maintenance would serve a frozen backfilled count;
deploying writes before the backfill would increment from 0. Hence the order above.

---

## Files

| Repo | File | Change |
|------|------|--------|
| eirka-post | `eirka.sql` | add `tag_count` column + index to `CREATE TABLE tags` |
| eirka-post | `migrations/tag_count.sql` (new) | ALTER + backfill for live DBs |
| eirka-post | `models/addtag.go` | `+1` inside the existing tx |
| eirka-admin | `models/deleteimagetag.go` | wrap in tx, `-1` |
| eirka-get | `models/tags.go` | swap to single-table read |
| eirka-get | `models/tagsearch.go` | use `tag_count` instead of correlated subquery |
| eirka-get | `models/tag.go` | use `tag_count` for header (confirm semantics) |

---

## Verification

- **Backfill correctness:** for several tags, `SELECT tag_count FROM tags WHERE tag_id = X`
  equals `SELECT COUNT(*) FROM tagmap WHERE tag_id = X`.
- **Maintenance:** add a tag to an image (addtag) → counter `+1`; remove it (deleteimagetag) →
  counter `-1`; re-check against `COUNT(*)`.
- **Read endpoint:** `GET /tags/:ib/1` returns the same JSON shape and popularity order as
  before, now backed by `tag_count`. Compare top-N ordering against the old query on a data
  copy.
- **Performance:** `EXPLAIN` the new `tags.go` query — expect a `ref` on `ib_id` using
  `idx_tags_ibid_count`, ordering served by the index, no `Using filesort` over `tagmap`, no
  derived table.
- **Tests:** update `models/tags_test.go` (and tagsearch/tag tests) to mock the new
  single-table query; run `go fmt ./...`, `go vet ./...`, `go test ./...` in each repo. Keep no
  skipped/stub tests.

---

## Considered alternatives (not chosen)

- **Periodic summary table `tag_counts(tag_id, count)`** refreshed by a MySQL event/cron —
  fully DB-side, no cross-repo app changes, but eventually-consistent and still requires a
  filesort over a board's tags on read. Made unnecessary because raw-count semantics let us
  maintain an *exact* counter cheaply at just two write sites.
- **Accurate (non-deleted) denormalized counter** — would require intercepting post/thread
  soft-delete toggles and re-aggregating affected tags at delete time (or DB triggers). Far
  more complex; rejected once raw-count semantics were accepted.
- **Pure query/index tuning only** — this is Phase 0; it improves the constant factor but stays
  O(all board tags) per cache miss, so it isn't a durable fix on its own.
