create database if not exists golang_jwt_auth;

use golang_jwt_auth;

create table if not exists users (
    id bigint primary key AUTO_INCREMENT,
    email varchar(255) NOT NULL UNIQUE,
    password varchar(255) NOT NULL,
    created_at timestamp default current_timestamp
);

create table if not exists refresh_tokens_table (
    user_id bigint UNIQUE,
    refresh_token varchar(255) NOT NULL,
    expire_time bigint NOT NULL,
    created_at timestamp default CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);