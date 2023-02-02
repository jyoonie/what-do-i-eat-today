-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.users
(
    user_uuid uuid not null default gen_random_uuid()
        constraint users_primary_key
            primary key,
    hashed_password varchar(128)    not null,
    active          boolean         not null default true,
    first_name      varchar(64)     not null,
    last_name       varchar(64)     not null,
    email_address   varchar(128)    not null UNIQUE,
    created_at      timestamp       not null default now(),
    updated_at      timestamp       not null default now()
);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.users;
-- +goose StatementEnd

-- +goose StatementBegin
DROP SCHEMA IF EXISTS wdiet;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.ingredients;
(
    ingredient_uuid uuid not null default gen_random_uuid()
        constraint ingredients_primary_key
            primary key,
    ingredient_name      varchar(64)     not null UNIQUE,
    category             varchar(64)     not null,
    days_until_exp       integer         not null,
    created_at           timestamp       not null default now(),
    updated_at           timestamp       not null default now()
);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.ingredients;
-- +goose StatementEnd

-- +goose StatementBegin
DROP SCHEMA IF EXISTS wdiet;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.fridges; --change the way the data looks.
(
    fridge_uuid uuid not null default gen_random_uuid()
        constraint fridges_primary_key
            primary key,
    user_uuid        uuid            not null UNIQUE,
    fridge_name      varchar(64)     not null UNIQUE,
    created_at       timestamp       not null default now(),
    updated_at       timestamp       not null default now()
);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.fridges;
-- +goose StatementEnd

-- +goose StatementBegin
DROP SCHEMA IF EXISTS wdiet;
-- +goose StatementEnd