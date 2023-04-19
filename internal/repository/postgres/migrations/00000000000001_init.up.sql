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
    is_activated boolean,
    UNIQUE (id, telephone, telegram_username, email)
);

CREATE TYPE event AS ENUM ('proposal-event', 'help', 'public');
CREATE type event_status AS ENUM('active', 'inactive', 'done', 'blocked');

CREATE TABLE propositional_event (
    id bigserial PRIMARY KEY,
    title varchar,
    description varchar,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    competition_date timestamp,
    status event_status DEFAULT 'inactive',
    max_concurrent_requests integer NOT NULL,
    remaining_helps integer NOT NULL,
    image_path varchar,
    author_id bigint,
    is_deleted bool,
    CONSTRAINT author_fk FOREIGN KEY(author_id) REFERENCES members(id)
);

CREATE TYPE priority_rate AS ENUM('normal', 'fast', 'very fast');

CREATE TABLE help_event (
    id bigserial PRIMARY KEY,
    title varchar,
    description varchar,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author_id bigint,
    category varchar,
    competition_date timestamp,
    rate priority_rate DEFAULT 'normal' NOT NULL,
    CONSTRAINT author_fk FOREIGN KEY(author_id) REFERENCES members(id)
);


CREATE TABLE IF NOT EXISTS tag (
    id bigserial PRIMARY KEY,
    title varchar(255),
    event_id bigint,
    event_type event
);

CREATE TABLE IF NOT EXISTS tag_value (
    id bigserial PRIMARY KEY,
    tag_id bigint,
    value varchar(255),
    CONSTRAINT tag_id FOREIGN KEY(tag_id) REFERENCES tag(id)
        ON DELETE CASCADE ON UPDATE CASCADE
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

CREATE TABLE comment (
    id bigserial PRIMARY KEY,
    event_id bigint,
    event_type event,
    text varchar(255),
    user_id bigint,
    updated_at timestamp,
    is_updated bool,
    is_deleted bool,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE type responder_status AS ENUM ('not_started', 'in_progress', 'completed', 'aborted');
CREATE type transaction_status AS ENUM ('waiting', 'in_progress', 'completed', 'aborted', 'canceled', 'accepted', 'waiting_for_approve');

CREATE TABLE transaction (
    id bigserial PRIMARY KEY,
    creator_id bigint,
    creation_date timestamp DEFAULT now() NOT NULL,
    completion_date timestamp,
    event_id bigint,
    comment varchar(255),
    event_type event,
    transaction_status transaction_status DEFAULT 'waiting' NOT NULL,
    responder_status responder_status DEFAULT 'not_started' NOT NULL,
    report_url varchar,
    CONSTRAINT creator_fk FOREIGN KEY(creator_id) REFERENCES members(id)
        ON DELETE SET NULL ON UPDATE SET DEFAULT
);

END;