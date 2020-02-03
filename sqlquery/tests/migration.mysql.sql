CREATE TABLE `robots` (
  `id`   int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(11)               DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `posts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `users` (
  `id`              int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name`            varchar(100)              DEFAULT 'Wall-E',
  `surname`         varchar(100)              DEFAULT NULL,
  `age`             int(11)                   DEFAULT NULL,
  `birthday`        date                      DEFAULT NULL,
  `deleted_at`      datetime                  DEFAULT NULL,
  `type_int`        int(11)                   DEFAULT NULL,
  `type_uint`       int(10) unsigned          DEFAULT NULL,
  `type_bigint`     bigint(11)                DEFAULT NULL,
  `type_ubigint`    bigint(20) unsigned       DEFAULT NULL,
  `type_mediumint`  mediumint(9)              DEFAULT NULL,
  `type_umediumint` mediumint(9) unsigned     DEFAULT NULL,
  `type_smallint`   smallint(6)               DEFAULT NULL,
  `type_usmallint`  smallint(6) unsigned      DEFAULT NULL,
  `type_tinyint`    tinyint(4)                DEFAULT NULL,
  `type_utinyint`   tinyint(4) unsigned       DEFAULT NULL,
  `type_float`      float                     DEFAULT NULL,
  `type_double`     double                    DEFAULT NULL,
  `type_varchar`    varchar(10)               DEFAULT NULL,
  `type_char`       char(10)                  DEFAULT NULL,
  `type_tinytext`   tinytext,
  `type_text`       text,
  `type_mediumtext` mediumtext,
  `type_longtext`   longtext,
  `type_time`       time                      DEFAULT NULL,
  `type_date`       date                      DEFAULT NULL,
  `type_timestamp`  timestamp        NULL     DEFAULT NULL,
  `type_datetime`   datetime                  DEFAULT NULL,
  `type_geometry`   geometry                  DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `user_posts` (
  `user_id` int(11) unsigned NOT NULL,
  `post_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `post_id`),
  KEY `post_id` (`post_id`),
  CONSTRAINT `user_posts_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `user_posts_ibfk_2` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `addresses` (
  `id`      int(11) unsigned NOT NULL AUTO_INCREMENT,
  `street`  varchar(11)               DEFAULT NULL,
  `zip`     varchar(10)               DEFAULT NULL,
  `country` varchar(2)                DEFAULT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `addresses_ibfk_1` FOREIGN KEY (`id`) REFERENCES `users` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `histories` (
  `id`      int(11) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) unsigned NOT NULL,
  `text`    text             NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `histories_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;