CREATE SEQUENCE robots_seq;
CREATE TABLE robots (
  id   int check (id > 0) NOT NULL DEFAULT NEXTVAL('robots_seq'),
  name varchar(11)                 DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE SEQUENCE users_seq;
CREATE TABLE users
(
  id                integer NOT NULL                                    DEFAULT nextval('users_seq'::regclass),
  name              character varying(100) COLLATE pg_catalog."default" DEFAULT 'Wall-E'::character varying,
  surname           character varying(100) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
  age               integer,
  birthday          date,
  deleted_at        timestamp(0) without time zone DEFAULT NULL ::timestamp without time zone,
  type_int          integer,
  type_bigint       bigint,
  type_smallint     smallint,
  type_real         real,
  type_double       double precision,
  type_numeric      numeric,
  type_varchar      character varying(10) COLLATE pg_catalog."default"  DEFAULT NULL::character varying,
  type_char         character(10) COLLATE pg_catalog."default"          DEFAULT NULL::bpchar,
  type_text         text COLLATE pg_catalog."default",
  type_date         date,
  type_timestamp_tz timestamp(6) with time zone DEFAULT NULL ::timestamp without time zone,
  type_timestamp    timestamp(0) without time zone DEFAULT NULL ::timestamp without time zone,
  type_time         time(6) without time zone,
  type_time_tz      time with time zone,
  type_uuid uuid,
  CONSTRAINT users_pkey PRIMARY KEY (id),
  CONSTRAINT users_id_check CHECK (id > 0)
);

CREATE SEQUENCE addresses_seq;
CREATE TABLE addresses
(
  id      integer                                            NOT NULL DEFAULT nextval('addresses_seq'::regclass),
  street  character varying(11) COLLATE pg_catalog."default" NULL,
  zip     character varying(10) COLLATE pg_catalog."default" NULL,
  country character varying(2) COLLATE pg_catalog."default"  NULL,
  CONSTRAINT addresses_pkey PRIMARY KEY (id),
  CONSTRAINT addresses_ibfk_1 FOREIGN KEY (id)
  REFERENCES users (id)
    MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION,
  CONSTRAINT addresses_id_check CHECK (id > 0)
);

CREATE SEQUENCE histories_seq;
CREATE TABLE histories
(
  id      integer                           NOT NULL DEFAULT nextval('histories_seq'::regclass),
  user_id integer                           NOT NULL,
  text    text COLLATE pg_catalog."default" NOT NULL,
  CONSTRAINT histories_pkey PRIMARY KEY (id),
  CONSTRAINT histories_ibfk_1 FOREIGN KEY (user_id)
  REFERENCES users (id)
    MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION,
  CONSTRAINT histories_id_check CHECK (id > 0),
  CONSTRAINT histories_user_id_check CHECK (user_id > 0)
);
CREATE INDEX user_id
  ON histories USING btree (user_id) TABLESPACE pg_default;


CREATE SEQUENCE posts_seq;
CREATE TABLE posts
(
  id integer NOT NULL DEFAULT nextval('posts_seq'::regclass),
  CONSTRAINT posts_pkey PRIMARY KEY (id),
  CONSTRAINT posts_id_check CHECK (id > 0)
);

CREATE TABLE user_posts
(
  user_id integer NOT NULL,
  post_id integer NOT NULL,
  CONSTRAINT user_posts_pkey PRIMARY KEY (user_id, post_id),
  CONSTRAINT user_posts_ibfk_1 FOREIGN KEY (user_id)
  REFERENCES users (id)
    MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION,
  CONSTRAINT user_posts_ibfk_2 FOREIGN KEY (post_id)
  REFERENCES posts (id)
    MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION,
  CONSTRAINT user_posts_user_id_check CHECK (user_id > 0),
  CONSTRAINT user_posts_post_id_check CHECK (post_id > 0)
);

CREATE INDEX post_id
  ON public.user_posts USING btree (post_id) TABLESPACE pg_default;
