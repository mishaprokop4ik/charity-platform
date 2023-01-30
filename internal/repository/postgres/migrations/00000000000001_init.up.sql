BEGIN;

CREATE TABLE members (
    id bigserial PRIMARY KEY,
    full_name varchar,
    telephone varchar(13),
    telegram_username varchar,
    email varchar,
    password varchar,
    image_path varchar,
    company_name varchar,
    address varchar,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp,
    deleted_at timestamp,
    is_deleted boolean,
    is_admin boolean,
    UNIQUE (id, telephone, telegram_username, email)
);

CREATE TYPE event AS ENUM ('propositional', 'help', 'public');

CREATE TABLE propositional_event (
    id bigserial PRIMARY KEY,
    title varchar,
    description varchar,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    competition_date timestamp,
    author_id bigint,
    category varchar,
    CONSTRAINT author_fk FOREIGN KEY(author_id) REFERENCES members(id)
);

CREATE TYPE priority_rate AS ENUM('нормально', 'швидко', 'дуже швидко');

CREATE TABLE help_event (
    id bigserial PRIMARY KEY,
    title varchar,
    description varchar,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author_id bigint,
    category varchar,
    address varchar,
    competition_date timestamp,
    rate priority_rate,
    CONSTRAINT author_fk FOREIGN KEY(author_id) REFERENCES members(id)
);

CREATE TABLE report (
    id bigserial PRIMARY KEY,
    s3_path varchar,
    event_type event,
    transaction_id bigint,
    members_id bigint,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT members_fk FOREIGN KEY(members_id) REFERENCES members(id)
);

END;