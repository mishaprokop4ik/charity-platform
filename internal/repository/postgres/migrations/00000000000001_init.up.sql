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
CREATE type event_status AS ENUM('active', 'inactive', 'done');

CREATE TABLE propositional_event (
    id bigserial PRIMARY KEY,
    title varchar,
    description varchar,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    competition_date timestamp,
    status event_status DEFAULT 'active',
    max_concurrent_requests integer NOT NULL,
    remaining_helps integer NOT NULL,
    author_id bigint,
    category varchar,
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
    address varchar,
    competition_date timestamp,
    rate priority_rate,
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
                     -- Add user and event transaction
);

CREATE type status AS ENUM ('in_process', 'completed', 'interrupted', 'canceled', 'waiting');

CREATE TABLE transaction (
    id bigserial PRIMARY KEY,
    creator_id bigint,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completion_date timestamp,
    event_id bigint,
    last_comment varchar(255),
    status status,
    event_type event,
    transaction_status status NOT NULL,
    responder_status status NOT NULL,
    CONSTRAINT creator_fk FOREIGN KEY(creator_id) REFERENCES members(id)
        ON DELETE SET NULL ON UPDATE SET DEFAULT
);

END;