
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

DROP TABLE IF EXISTS `audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `audit` (
  `audit_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL,
  `ib_id` tinyint(3) unsigned NOT NULL,
  `audit_ip` varchar(255) NOT NULL,
  `audit_time` datetime NOT NULL,
  `audit_action` varchar(255) NOT NULL,
  `audit_info` varchar(255) NOT NULL,
  PRIMARY KEY (`audit_id`),
  KEY `audit_user_id_idx` (`user_id`),
  KEY `audit_ib_id_idx` (`ib_id`),
  CONSTRAINT `audit_ib_id` FOREIGN KEY (`ib_id`) REFERENCES `imageboards` (`ib_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `audit_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `imageboards`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `imageboards` (
  `ib_id` tinyint(3) unsigned NOT NULL AUTO_INCREMENT,
  `ib_title` varchar(45) COLLATE utf8_unicode_ci NOT NULL,
  `ib_description` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `ib_domain` varchar(40) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`ib_id`),
  KEY `ib_id_ib_title` (`ib_id`,`ib_title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `images`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `images` (
  `image_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `post_id` int(10) unsigned NOT NULL,
  `image_deleted` tinyint(1) NOT NULL,
  `image_file` varchar(20) COLLATE utf8_unicode_ci NOT NULL,
  `image_thumbnail` varchar(20) COLLATE utf8_unicode_ci NOT NULL,
  `image_hash` varchar(32) COLLATE utf8_unicode_ci NOT NULL,
  `image_orig_height` smallint(5) unsigned NOT NULL DEFAULT '0',
  `image_orig_width` smallint(5) unsigned NOT NULL DEFAULT '0',
  `image_tn_height` smallint(5) unsigned NOT NULL DEFAULT '0',
  `image_tn_width` smallint(5) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`image_id`),
  KEY `post_id_idx` (`post_id`),
  KEY `p_id_i_id` (`post_id`,`image_id`),
  KEY `hash_idx` (`image_hash`),
  CONSTRAINT `post_id` FOREIGN KEY (`post_id`) REFERENCES `posts` (`post_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `posts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `posts` (
  `post_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `thread_id` smallint(5) unsigned NOT NULL,
  `post_name` varchar(20) COLLATE utf8_unicode_ci NOT NULL,
  `post_num` smallint(5) unsigned NOT NULL DEFAULT '1',
  `post_ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `post_time` datetime NOT NULL,
  `post_text` text COLLATE utf8_unicode_ci,
  PRIMARY KEY (`post_id`),
  KEY `thread_id_idx` (`thread_id`),
  KEY `t_id_p_id` (`thread_id`,`post_id`),
  CONSTRAINT `thread_id` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`thread_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `settings` (
  `settings_key` varchar(255) NOT NULL,
  `settings_value` varchar(255) NOT NULL,
  PRIMARY KEY (`settings_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `tagmap`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tagmap` (
  `image_id` int(10) unsigned NOT NULL,
  `tag_id` int(10) unsigned NOT NULL,
  `tagmap_ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`image_id`,`tag_id`),
  KEY `image_id_idx` (`image_id`),
  KEY `tag_id_idx` (`tag_id`),
  KEY `tagmap_tag_i` (`image_id`,`tag_id`),
  CONSTRAINT `image_id` FOREIGN KEY (`image_id`) REFERENCES `images` (`image_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`tag_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tags` (
  `tag_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ib_id` tinyint(3) unsigned NOT NULL,
  `tagtype_id` int(10) unsigned NOT NULL,
  `tag_name` varchar(128) COLLATE utf8_unicode_ci NOT NULL,
  `tag_ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`tag_id`),
  KEY `tags_ib_id_idx` (`ib_id`),
  KEY `tag_id_tag_name` (`tag_id`,`tag_name`),
  KEY `tagtype_id_idx` (`tagtype_id`),
  KEY `tt_t_id` (`tagtype_id`,`tag_id`),
  CONSTRAINT `ib_id_tags` FOREIGN KEY (`ib_id`) REFERENCES `imageboards` (`ib_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `tagtype_id` FOREIGN KEY (`tagtype_id`) REFERENCES `tagtype` (`tagtype_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `tagtype`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tagtype` (
  `tagtype_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tagtype_name` varchar(45) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`tagtype_id`),
  KEY `tt_id_tt_name` (`tagtype_id`,`tagtype_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `threads` (
  `thread_id` smallint(5) unsigned NOT NULL AUTO_INCREMENT,
  `ib_id` tinyint(3) unsigned NOT NULL,
  `thread_title` varchar(45) COLLATE utf8_unicode_ci NOT NULL,
  `thread_closed` tinyint(1) NOT NULL,
  `thread_sticky` tinyint(1) NOT NULL,
  `thread_deleted` tinyint(1) NOT NULL,
  `thread_first_post` datetime NOT NULL,
  `thread_last_post` datetime NOT NULL,
  PRIMARY KEY (`thread_id`),
  KEY `ib_id_idx` (`ib_id`),
  KEY `t_id_ib_id` (`ib_id`,`thread_id`),
  CONSTRAINT `ib_id` FOREIGN KEY (`ib_id`) REFERENCES `imageboards` (`ib_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `users` (
  `user_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_name` varchar(20) NOT NULL,
  `user_email` varchar(255) DEFAULT NULL,
  `password` char(60) DEFAULT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `settings` WRITE;
/*!40000 ALTER TABLE `settings` DISABLE KEYS */;
INSERT INTO `settings` VALUES ('antispam_cookiename','pram_antispam'),('antispam_cookievalue','not_a_bot'),('antispam_key','j5ebMACL'),('comment_maxlength','1000'),('comment_minlength','5'),('image_maxheight','20000'),('image_maxsize','20000000'),('image_maxwidth','20000'),('image_minheight','100'),('image_minwidth','100'),('index_postsperthread','6'),('index_threadsperpage','10'),('name_maxlength','20'),('name_minlength','3'),('param_maxsize','1000000'),('tag_maxlength','128'),('tag_minlength','3'),('thread_postsmax','800'),('thread_postsperpage','100'),('thumbnail_maxheight','300'),('thumbnail_maxwidth','200'),('title_maxlength','40'),('title_minlength','3'),('webm_maxlength','300');
/*!40000 ALTER TABLE `settings` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `imageboards` WRITE;
/*!40000 ALTER TABLE `imageboards` DISABLE KEYS */;
INSERT INTO `imageboards` VALUES (1,'Dev','Dev Board','dev.whatever.com');
/*!40000 ALTER TABLE `imageboards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `tagtype` WRITE;
/*!40000 ALTER TABLE `tagtype` DISABLE KEYS */;
INSERT INTO `tagtype` VALUES (1,'Tag'),(2,'Artist'),(3,'Character'),(4,'Copyright');
/*!40000 ALTER TABLE `tagtype` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `tags` WRITE;
/*!40000 ALTER TABLE `tags` DISABLE KEYS */;
INSERT INTO `tags` VALUES (1,1,1,'Default','0.0.0.0');
/*!40000 ALTER TABLE `tags` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'Anonymous',NULL,NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
