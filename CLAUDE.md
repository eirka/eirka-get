# CLAUDE.md - Commands & Guidelines for eirka-get

## Build Commands
- Build: `go build`
- Run tests: `go test ./...`
- Run specific test: `go test -v ./package/... -run TestName`
- Run tests with coverage: `go test ./... -cover`
- Get dependencies: `go mod tidy`

## Code Style Guidelines
- **Formatting**: Follow Go standard formatting (gofmt)
- **Imports**: Group imports in 3 blocks: standard lib, third-party, project-local
- **Error Handling**: Use error wrapping with `SetMeta()` to provide context
- **Naming**: 
  - Use camelCase for variables, PascalCase for exported functions
  - Short variable names for scope-limited vars
- **Error Pattern**: Return errors with `c.JSON(e.ErrorMessage())` + `c.Error().SetMeta()`
- **Testing**: Use `testify/assert` for test assertions
- **Controllers**: Return JSON responses, handle validation in middleware
- **Context**: Use Gin context to pass data between middleware/controllers

## Project Structure
- Models: Handle data access
- Controllers: HTTP handlers 
- Middleware: Request processing (analytics, cache, etc.)
- Config: Application configuration

## Cache Middleware
The cache middleware implements Redis caching with circuit breaker pattern:

- Uses Redis to cache API responses
- Implements circuit breaker to bypass cache when Redis is failing
- Uses singleflight to prevent duplicate database queries
- Default request timeout: 10 seconds (configurable in tests with `testTimeout`)
- Skips caching for URLs with query parameters
- Controllers must properly use the callback mechanism:
  ```go
  if callback, exists := c.Get("setDataCallback"); exists {
      callback.(func([]byte, error))(jsonData, nil)
  }
  ```

## Testing Guidelines
- Always set a short timeout in cache middleware tests:
  ```go
  c.Set("testTimeout", 100*time.Millisecond)
  ```
- Use Redis mocks for cache tests:
  ```go
  redis.NewRedisMock()
  redis.Cache.Mock.Command("HGET", "key", "value").Expect("result")
  ```
- Reset circuit breaker state before tests:
  ```go
  CircuitBreaker = NewCircuitBreaker()
  ```
- For concurrent tests, use WaitGroups and Mutexes to manage state

## Test prerequisites
Go toolchain only; no MySQL or Redis. 19 of 22 `_test.go` files use `gopkg.in/DATA-DOG/go-sqlmock.v1`; `middleware/cache_test.go` uses `redis.NewRedisMock()` (redigomock via eirka-libs). CI `.github/workflows/go.yml` pins setup-go `1.22.x` (go.mod: `go 1.26.0`); leave the CI pin alone on dep bumps.

## Fleet conventions
Siblings eirka-post, eirka-admin and eirka-index share `github.com/eirka/eirka-libs`; eirka-post owns the schema (`eirka.sql`), and cross-repo tag-count work follows the standing design in `docs/tag-count-performance.md`; `make check` = build + lint + test.
