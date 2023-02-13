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

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.ingredients
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

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.fridge_ingredients --change the way the data looks.
(
    user_uuid              uuid            not null
        constraint user_uuid_fk references wdiet.users,
    ingredient_uuid        uuid            not null
        constraint ingredient_uuid_fk references wdiet.ingredients,
    amount                 integer         not null,
    unit                   varchar(64)     not null,
    purchased_date         timestamp       not null, --expiry_date는 days until expire랑은 달라.. 그건 integer고 이건 date임. 근데 expiry_date도 하지마. ingredient의 days until exp를 이용해서 자동계산하려면 구매한 날짜만 알면됨. 그럼 유통기한은 자동 계산되니까.
    expiration_date        timestamp       not null,
    created_at             timestamp       not null default now(),
    updated_at             timestamp       not null default now(),
    PRIMARY KEY (user_uuid, ingredient_uuid) --PRIMARY KEY doesn't need CREATE INDEX ON
);

-- CREATE INDEX ON wdiets.fridge_ingredients (user_uuid);
-- CREATE INDEX ON wdiets.fridge_ingredients (ingredient_uuid);

-- ALTER TABLE wdiet.fridge_ingredients ADD FOREIGN KEY user_uuid REFERENCES wdiet.users (user_uuid);
-- ALTER TABLE wdiet.fridge_ingredients ADD FOREIGN KEY ingredient_uuid REFERENCES wdiet.ingredients (ingredient_uuid);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.recipes
(
    recipe_uuid uuid not null default gen_random_uuid()
        constraint recipes_primary_key
            primary key,
    user_uuid            uuid            not null
        constraint user_uuid_fk references wdiet.users,
    recipe_name          varchar(64)     not null,
    category             varchar(64)     not null,
    created_at           timestamp       not null default now(),
    updated_at           timestamp       not null default now()
);

CREATE INDEX ON wdiet.recipes (user_uuid);
-- ALTER TABLE wdiet.recipes ADD FOREIGN KEY user_uuid REFERENCES wdiet.users (user_uuid);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.recipe_ingredients
(
    recipe_uuid            uuid            not null
        constraint recipe_uuid_fk references wdiet.recipes,
    ingredient_uuid        uuid            not null
        constraint ingredient_uuid_fk references wdiet.ingredients,
    amount                 integer         not null,
    unit                   varchar(64)     not null,
    PRIMARY KEY (recipe_uuid, ingredient_uuid)
    -- created_at             timestamp       not null default now(), --레시피에 있는 필드니까 안해줘도 상관없음.
    -- updated_at             timestamp       not null default now()
);

-- CREATE INDEX ON wdiets.recipe_ingredients (recipe_uuid);
-- ALTER TABLE wdiet.recipe_ingredients ADD FOREIGN KEY recipe_uuid REFERENCES wdiet.recipes (recipe_uuid);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS wdiet;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wdiet.recipe_instructions
(
    recipe_uuid            uuid            not null
        constraint recipe_uuid_fk references wdiet.recipes,
    step_num               integer         not null,
    instruction            varchar(1024)   not null,
    PRIMARY KEY (recipe_uuid, step_num) --하나의 레시피에 한가지의 단계만 guarantee하기 위함. 1. 양파를 썬다. 1. 육수를 낸다. 1. 고기를 썬다. 이러면 안되니까. 1, 2, 3단계여야 하니까.
    -- created_at             timestamp       not null default now(),
    -- updated_at             timestamp       not null default now()
);

-- CREATE INDEX ON wdiets.recipe_instructions (recipe_uuid); efficient to search the table on this column. Binary tree.
-- ALTER TABLE wdiet.recipe_instructions ADD FOREIGN KEY recipe_uuid REFERENCES wdiet.recipes (recipe_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.fridge_ingredients;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.recipe_ingredients;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.recipe_instructions;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.ingredients;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.recipes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS wdiet.users;
-- +goose StatementEnd

-- +goose StatementBegin
DROP SCHEMA IF EXISTS wdiet;
-- +goose StatementEnd