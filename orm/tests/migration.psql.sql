--
-- Dump of table accountfks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS accountfks;

CREATE SEQUENCE accountfks_seq;

CREATE TABLE accountfks (
  id   int check (id > 0) NOT NULL DEFAULT NEXTVAL('accountfks_seq'),
  name varchar(100)                DEFAULT NULL,
  PRIMARY KEY (id)
);

-- Dump of table customerfks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS customerfks;

CREATE SEQUENCE customerfks_seq;

CREATE TABLE customerfks (
  id         int check (id > 0)         NOT NULL DEFAULT NEXTVAL('customerfks_seq'),
  first_name varchar(120)                        DEFAULT NULL,
  last_name  varchar(120)                        DEFAULT NULL,
  created_at timestamp(0)               NULL     DEFAULT NULL,
  updated_at timestamp(0)               NULL     DEFAULT NULL,
  deleted_at timestamp(0)               NULL     DEFAULT NULL,
  account_id int check (account_id > 0) NOT NULL,
  PRIMARY KEY (id)
  ,
  CONSTRAINT customerfks_ibfk_1 FOREIGN KEY (account_id) REFERENCES accountfks (id)
);

CREATE INDEX account_id
  ON customerfks (account_id);

-- Dump of table contactfks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS contactfks;

CREATE SEQUENCE contactfks_seq;

CREATE TABLE contactfks (
  id          int check (id > 0)          NOT NULL DEFAULT NEXTVAL('contactfks_seq'),
  customer_id int check (customer_id > 0) NOT NULL,
  phone       varchar(120)                         DEFAULT NULL,
  PRIMARY KEY (id)
  ,
  CONSTRAINT contactfks_ibfk_1 FOREIGN KEY (customer_id) REFERENCES customerfks (id)
);

CREATE INDEX INDEXcustomer_id1
  ON contactfks (customer_id);

-- Dump of table contacts
-- ------------------------------------------------------------

DROP TABLE IF EXISTS contacts;

CREATE SEQUENCE contacts_seq;

CREATE TABLE contacts (
  id          int check (id > 0)          NOT NULL DEFAULT NEXTVAL('contacts_seq'),
  customer_id int check (customer_id > 0) NOT NULL,
  phone       varchar(120)                         DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX INDEXcustomer_id2
  ON contacts (customer_id);

-- Dump of table servicefks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS servicefks;

CREATE SEQUENCE servicefks_seq;

CREATE TABLE servicefks (
  id   int check (id > 0) NOT NULL DEFAULT NEXTVAL('servicefks_seq'),
  name varchar(120)                DEFAULT NULL,
  PRIMARY KEY (id)
);

-- Dump of table services
-- ------------------------------------------------------------

DROP TABLE IF EXISTS services;

CREATE SEQUENCE services_seq;

CREATE TABLE services (
  id   int check (id > 0) NOT NULL DEFAULT NEXTVAL('services_seq'),
  name varchar(120)                DEFAULT NULL,
  PRIMARY KEY (id)
);

-- Dump of table customerfk_servicefks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS customerfk_servicefks;

CREATE TABLE customerfk_servicefks (
  customer_id int check (customer_id > 0) NOT NULL,
  service_id  int check (service_id > 0)  NOT NULL,
  PRIMARY KEY (customer_id, service_id)
  ,
  CONSTRAINT customerfk_servicefks_ibfk_1 FOREIGN KEY (customer_id) REFERENCES customerfks (id),
  CONSTRAINT customerfk_servicefks_ibfk_2 FOREIGN KEY (service_id) REFERENCES servicefks (id)
);

CREATE INDEX service_id
  ON customerfk_servicefks (service_id);

-- Dump of table customers
-- ------------------------------------------------------------

DROP TABLE IF EXISTS customers;

CREATE SEQUENCE customers_seq;

CREATE TABLE customers (
  id         int check (id > 0) NOT NULL DEFAULT NEXTVAL('customers_seq'),
  first_name varchar(120)                DEFAULT NULL,
  last_name  varchar(120)                DEFAULT NULL,
  created_at timestamp(0)       NULL     DEFAULT NULL,
  updated_at timestamp(0)       NULL     DEFAULT NULL,
  deleted_at timestamp(0)       NULL     DEFAULT NULL,
  PRIMARY KEY (id)
);

-- Dump of table orderfks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS orderfks;

CREATE SEQUENCE orderfks_seq;

CREATE TABLE orderfks (
  id          int check (id > 0)          NOT NULL DEFAULT NEXTVAL('orderfks_seq'),
  customer_id int check (customer_id > 0) NOT NULL,
  created_at  timestamp(0)                         DEFAULT NULL,
  PRIMARY KEY (id)
  ,
  CONSTRAINT orderfks_ibfk_1 FOREIGN KEY (customer_id) REFERENCES customerfks (id)
    ON DELETE CASCADE
);

CREATE INDEX INDEXcustomer_id3
  ON orderfks (customer_id);

-- Dump of table orders
-- ------------------------------------------------------------

DROP TABLE IF EXISTS orders;

CREATE SEQUENCE orders_seq;

CREATE TABLE orders (
  id          int check (id > 0)          NOT NULL DEFAULT NEXTVAL('orders_seq'),
  customer_id int check (customer_id > 0) NOT NULL,
  created_at  timestamp(0)                         DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX INDEXcustomer_id4
  ON orders (customer_id);

-- Dump of table productfks
-- ------------------------------------------------------------

DROP TABLE IF EXISTS productfks;

CREATE SEQUENCE productfks_seq;

CREATE TABLE productfks (
  id         int check (id > 0)       NOT NULL DEFAULT NEXTVAL('productfks_seq'),
  name       varchar(120)                      DEFAULT NULL,
  price      double precision                  DEFAULT NULL,
  created_at timestamp(0)                      DEFAULT NULL,
  updated_at timestamp(0)                      DEFAULT NULL,
  order_id   int check (order_id > 0) NOT NULL,
  PRIMARY KEY (id)
  ,
  CONSTRAINT productfks_ibfk_1 FOREIGN KEY (order_id) REFERENCES orderfks (id)
);

CREATE INDEX order_id
  ON productfks (order_id);

-- Dump of table products
-- ------------------------------------------------------------

DROP TABLE IF EXISTS products;

CREATE SEQUENCE products_seq;

CREATE TABLE products (
  id         int check (id > 0)       NOT NULL DEFAULT NEXTVAL('products_seq'),
  name       varchar(120)                      DEFAULT NULL,
  price      double precision                  DEFAULT NULL,
  created_at int                               DEFAULT NULL,
  updated_at int                               DEFAULT NULL,
  order_id   int check (order_id > 0) NOT NULL,
  PRIMARY KEY (id)
);

