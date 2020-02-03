#
# Dump of table accountfks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `accountfks`;

CREATE TABLE `accountfks` (
  `id`   int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100)              DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table customerfks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `customerfks`;

CREATE TABLE `customerfks` (
  `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
  `first_name` varchar(120)              DEFAULT NULL,
  `last_name`  varchar(120)              DEFAULT NULL,
  `created_at` timestamp        NULL     DEFAULT NULL,
  `updated_at` timestamp        NULL     DEFAULT NULL,
  `deleted_at` timestamp        NULL     DEFAULT NULL,
  `account_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `customerfks_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `accountfks` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table contactfks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `contactfks`;

CREATE TABLE `contactfks` (
  `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` int(10) unsigned NOT NULL,
  `phone`       varchar(120)              DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `customer_id` (`customer_id`),
  CONSTRAINT `contactfks_ibfk_1` FOREIGN KEY (`customer_id`) REFERENCES `customerfks` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table contacts
# ------------------------------------------------------------

DROP TABLE IF EXISTS `contacts`;

CREATE TABLE `contacts` (
  `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` int(10) unsigned NOT NULL,
  `phone`       varchar(120)              DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `customer_id` (`customer_id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table servicefks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `servicefks`;

CREATE TABLE `servicefks` (
  `id`   int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(120)              DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table services
# ------------------------------------------------------------

DROP TABLE IF EXISTS `services`;

CREATE TABLE `services` (
  `id`   int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(120)              DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table customerfk_servicefks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `customerfk_servicefks`;

CREATE TABLE `customerfk_servicefks` (
  `customer_id` int(11) unsigned NOT NULL,
  `service_id`  int(11) unsigned NOT NULL,
  PRIMARY KEY (`customer_id`, `service_id`),
  KEY `service_id` (`service_id`),
  CONSTRAINT `customerfk_servicefks_ibfk_1` FOREIGN KEY (`customer_id`) REFERENCES `customerfks` (`id`),
  CONSTRAINT `customerfk_servicefks_ibfk_2` FOREIGN KEY (`service_id`) REFERENCES `servicefks` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table customers
# ------------------------------------------------------------

DROP TABLE IF EXISTS `customers`;

CREATE TABLE `customers` (
  `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
  `first_name` varchar(120)              DEFAULT NULL,
  `last_name`  varchar(120)              DEFAULT NULL,
  `created_at` timestamp        NULL     DEFAULT NULL,
  `updated_at` timestamp        NULL     DEFAULT NULL,
  `deleted_at` timestamp        NULL     DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table orderfks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `orderfks`;

CREATE TABLE `orderfks` (
  `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` int(10) unsigned NOT NULL,
  `created_at`  datetime                  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `customer_id` (`customer_id`),
  CONSTRAINT `orderfks_ibfk_1` FOREIGN KEY (`customer_id`) REFERENCES `customerfks` (`id`)
    ON DELETE CASCADE
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table orders
# ------------------------------------------------------------

DROP TABLE IF EXISTS `orders`;

CREATE TABLE `orders` (
  `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` int(10) unsigned NOT NULL,
  `created_at`  datetime                  DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `customer_id` (`customer_id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table productfks
# ------------------------------------------------------------

DROP TABLE IF EXISTS `productfks`;

CREATE TABLE `productfks` (
  `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name`       varchar(120)              DEFAULT NULL,
  `price`      float                     DEFAULT NULL,
  `created_at` datetime                  DEFAULT NULL,
  `updated_at` datetime                  DEFAULT NULL,
  `order_id`   int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `order_id` (`order_id`),
  CONSTRAINT `productfks_ibfk_1` FOREIGN KEY (`order_id`) REFERENCES `orderfks` (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

# Dump of table products
# ------------------------------------------------------------

DROP TABLE IF EXISTS `products`;

CREATE TABLE `products` (
  `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name`       varchar(120)              DEFAULT NULL,
  `price`      float                     DEFAULT NULL,
  `created_at` int(11)                   DEFAULT NULL,
  `updated_at` int(11)                   DEFAULT NULL,
  `order_id`   int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

