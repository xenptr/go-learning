-- create-tables.sql
-- Run this script from the MySQL CLI to bootstrap the "recordings" database.
--
--   mysql> source /path/to/create-tables.sql
--
-- It is safe to run multiple times: the DROP statement removes the old table
-- first so you always start from a clean slate.

DROP TABLE IF EXISTS album;

-- album holds data about vintage jazz recordings on vinyl.
-- id is an AUTO_INCREMENT primary key – MySQL assigns it automatically.
CREATE TABLE album (
  id         INT AUTO_INCREMENT NOT NULL,
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  price      DECIMAL(5,2) NOT NULL,
  PRIMARY KEY (`id`)
);

-- Seed the table with four rows of sample data.
INSERT INTO album
  (title, artist, price) VALUES
  ('Blue Train',    'John Coltrane',  56.99),
  ('Giant Steps',   'John Coltrane',  63.99),
  ('Jeru',          'Gerry Mulligan', 17.99),
  ('Sarah Vaughan', 'Sarah Vaughan',  34.98);
