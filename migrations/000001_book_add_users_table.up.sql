CREATE TABLE books(
id serial PRIMARY KEY,
title VARCHAR (50) UNIQUE NOT NULL,
year_book INTEGER NOT NULL,
user_id int NOT NULL
);

CREATE TABLE users(
user_id serial PRIMARY KEY,
email text unique not null,
password text not null
);

CREATE TABLE  sessions(
id serial PRIMARY KEY,
user_id int references users not null,
token text not null,
ip text not null,
user_agent text not null,
created_at timestamp not null default now()
);