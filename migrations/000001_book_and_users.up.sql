CREATE TABLE books(
id serial PRIMARY KEY,
title VARCHAR (50) UNIQUE NOT NULL,
year_book INTEGER NOT NULL
);

CREATE TABLE users
(
user_id serial PRIMARY KEY,
username text unique not null,
password text not null,
email text unique not null
);