BEGIN;
SET TIME ZONE 'Europe/Kiev';
SET timezone TO 'Europe/Kiev';
CREATE TABLE IF NOT EXISTS members
(
    id                bigserial PRIMARY KEY,
    full_name         varchar,
    telephone         varchar(15),
    telegram_username varchar,
    email             varchar,
    password          varchar,
    is_blocked        BOOLEAN   DEFAULT false             NOT NULL,
    image_path        varchar,
    company_name      varchar,
    address           varchar,
    created_at        timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at        timestamp,
    deleted_at        timestamp,
    is_deleted        boolean,
    is_admin          boolean,
    confirm_code integer[],
    is_activated      boolean,
    UNIQUE (email)
);

DO
$$
    BEGIN

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'need_unit') THEN
            CREATE TYPE need_unit AS ENUM
                (
                    'kilogram', 'liter', 'item', 'work'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'event') THEN
            CREATE TYPE event AS ENUM
                (
                    'proposal-event', 'help', 'public'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'event_status') THEN
            CREATE TYPE event_status AS ENUM
                (
                    'active', 'inactive', 'done', 'blocked'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'priority_rate') THEN
            CREATE TYPE priority_rate AS ENUM
                (
                    'normal', 'fast', 'very fast'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'priority_rate') THEN
            CREATE TYPE priority_rate AS ENUM
                (
                    'normal', 'fast', 'very fast'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'responder_status') THEN
            CREATE TYPE responder_status AS ENUM
                (
                    'not_started', 'in_progress', 'completed', 'aborted'
                    );
        END IF;

        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'transaction_status') THEN
            CREATE TYPE transaction_status AS ENUM
                (
                    'waiting', 'in_progress', 'completed', 'aborted', 'canceled', 'accepted', 'waiting_for_approve'
                    );
        END IF;

    END
$$;

CREATE TABLE IF NOT EXISTS propositional_event
(
    id                      bigserial PRIMARY KEY,
    title                   varchar,
    is_banned               BOOLEAN      DEFAULT false             NOT NULL,
    description             varchar,
    creation_date           timestamp    DEFAULT CURRENT_TIMESTAMP NOT NULL,
    competition_date        timestamp,
    status                  event_status DEFAULT 'inactive',
    max_concurrent_requests integer                                NOT NULL,
    remaining_helps         integer                                NOT NULL,
    image_path              varchar,
    author_id               bigint,
    end_date                timestamp,
    is_deleted              bool,
    CONSTRAINT author_fk FOREIGN KEY (author_id) REFERENCES members (id)
);

CREATE TABLE IF NOT EXISTS help_event
(
    id              bigserial PRIMARY KEY,
    title           varchar,
    description     varchar,
    creation_date   timestamp    DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by      bigint,
    status          event_status DEFAULT 'inactive'        NOT NULL,
    image_path      varchar,
    completion_time timestamp,
    is_banned       BOOLEAN      DEFAULT false             NOT NULL,
    is_deleted      BOOLEAN      DEFAULT false             NOT NULL,
    end_date        timestamp,
    CONSTRAINT author_fk FOREIGN KEY (created_by) REFERENCES members (id)
);


CREATE TABLE IF NOT EXISTS tag
(
    id
               bigserial
        PRIMARY
            KEY,
    title
               varchar(255),
    event_id   bigint,
    event_type event
);

CREATE TABLE IF NOT EXISTS tag_value
(
    id
        bigserial
        PRIMARY
            KEY,
    tag_id
        bigint,
    value
        varchar(255),
    CONSTRAINT tag_id FOREIGN KEY
        (
         tag_id
            ) REFERENCES tag
            (
             id
                )
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS report
(
    id             bigserial PRIMARY KEY,
    s3_path        varchar,
    event_type     event,
    transaction_id bigint,
    members_id     bigint,
    creation_date  timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT members_fk FOREIGN KEY (members_id) REFERENCES members (id)
);

CREATE TABLE IF NOT EXISTS comment
(
    id            bigserial PRIMARY KEY,
    event_id      bigint,
    event_type    event,
    text          varchar(255),
    user_id       bigint,
    updated_at    timestamp,
    is_updated    bool,
    is_deleted    bool,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction
(
    id                 bigserial PRIMARY KEY,
    creator_id         bigint,
    creation_date      timestamp          DEFAULT now()         NOT NULL,
    completion_date    timestamp,
    event_id           bigint,
    comment            varchar(255),
    event_type         event,
    transaction_status transaction_status DEFAULT 'waiting'     NOT NULL,
    responder_status   responder_status   DEFAULT 'not_started' NOT NULL,
    report_url         varchar,
    CONSTRAINT creator_fk FOREIGN KEY (creator_id) REFERENCES members (id)
        ON DELETE SET NULL ON UPDATE SET DEFAULT
);


CREATE TABLE IF NOT EXISTS location
(
    id         bigserial PRIMARY KEY,
    country    varchar default 'Ukraine' NOT NULL,
    area       varchar,
    city       varchar,
    district   varchar,
    street     varchar,
    home       varchar,
    event_type event,
    event_id   bigint
);
CREATE TABLE IF NOT EXISTS member_session
(
    refresh_token varchar,
    member_id     bigint,
    expires_at    timestamp NOT NULL,
    PRIMARY KEY (member_id),
    CONSTRAINT members_fk FOREIGN KEY (member_id) REFERENCES members (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);



CREATE TABLE IF NOT EXISTS member_search
(
    id         bigserial PRIMARY KEY,
    title      varchar(255),
    member_id  bigint,
    event_type event NOT NULL,
    CONSTRAINT member_id FOREIGN KEY (member_id) REFERENCES members (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE TABLE IF NOT EXISTS member_search_value
(
    id               bigserial PRIMARY KEY,
    member_search_id bigint,
    value            varchar(255),
    CONSTRAINT member_search_id FOREIGN KEY (member_search_id) REFERENCES member_search (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE TABLE IF NOT EXISTS notification
(
    id             BIGSERIAL PRIMARY KEY,
    event_type     event                   NOT NULL,
    event_id       BIGINT                  NOT NULL,
    action         varchar                 NOT NULL,
    transaction_id INTEGER                 NOT NULL,
    new_status     varchar                 NOT NULL,
    is_read        BOOLEAN   DEFAULT false NOT NULL,
    creation_time  TIMESTAMP DEFAULT now() NOT NULL,
    member_id      BIGINT                  NOT NULL
);

-- comments
CREATE OR REPLACE FUNCTION check_event_exists()
    RETURNS TRIGGER AS
$$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS(SELECT 1 FROM propositional_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS(SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER comment_event_exists_trigger
    BEFORE INSERT OR UPDATE
    ON comment
    FOR EACH ROW
EXECUTE FUNCTION check_event_exists();
-- transactions
CREATE OR REPLACE FUNCTION transaction_check_event_exists()
    RETURNS TRIGGER AS
$$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS(SELECT 1 FROM propositional_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS(SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER transaction_event_exists_trigger
    BEFORE INSERT OR UPDATE
    ON transaction
    FOR EACH ROW
EXECUTE FUNCTION transaction_check_event_exists();
-- notification
CREATE OR REPLACE FUNCTION notification_check_event_exists()
    RETURNS TRIGGER AS
$$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS(SELECT 1 FROM propositional_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS(SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER notification_event_exists_trigger
    BEFORE INSERT OR UPDATE
    ON notification
    FOR EACH ROW
EXECUTE FUNCTION notification_check_event_exists();

CREATE TABLE IF NOT EXISTS need
(
    id             bigserial PRIMARY KEY,
    title          varchar   NOT NULL,
    amount         integer,
    unit           need_unit NOT NULL,
    help_event_id  bigint,
    received       integer,
    received_total integer,
    transaction_id bigint,
    CONSTRAINT event_fk FOREIGN KEY (help_event_id) REFERENCES help_event (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT transaction_fk FOREIGN KEY (transaction_id) REFERENCES transaction (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS complaints
(
    id            bigserial PRIMARY KEY,
    description   varchar,
    event_type    event,
    created_by    bigint,
    event_id      bigint,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT author_fk FOREIGN KEY (created_by) REFERENCES members (id)
);
END;

BEGIN;

INSERT INTO members  (full_name, telephone, telegram_username, email, password, image_path, company_name, address,
                     created_at, updated_at, deleted_at, is_deleted, is_admin, is_activated)
VALUES ('John Doee', '+380671234567', '@john_doe', 'john.doe@example.com',
        '736c6461666a6173646e666d61736e88ea39439e74fa27c09a4fc0bc8ebe6d00978392',
        'https://charity-platform.s3.amazonaws.com/images/png-transparent-default-avatar-thumbnail.png', 'ABC Company',
        '123 Main St, Anytown, USA', '2022-01-01 00:00:00', null, null, false, true, true),
       ('Jane Smith', '+380671234567', '@jane_smith', 'jane.smith@example.com',
        '736c6461666a6173646e666d61736e88ea39439e74fa27c09a4fc0bc8ebe6d00978392',
        'https://charity-platform.s3.amazonaws.com/images/png-transparent-default-avatar-thumbnail.png',
        'XYZ Corporation', '456 Elm St, Anytown, USA', '2022-01-02 00:00:00', null, null, false, false, true),
       ('Bob Johnson', '+380671234567', '@bob_johnson', 'bob.johnson@example.com',
        '736c6461666a6173646e666d61736e88ea39439e74fa27c09a4fc0bc8ebe6d00978392',
        'https://charity-platform.s3.amazonaws.com/images/png-transparent-default-avatar-thumbnail.png', '123 LLC',
        '789 Oak St, Anytown, USA', '2022-01-03 00:00:00', null, null, false, false, true),
       ('Test Test', '+380671234567', '@bob_johnson', 'test@test.com',
        '736c6461666a6173646e666d61736e88ea39439e74fa27c09a4fc0bc8ebe6d00978392',
        'https://charity-platform.s3.amazonaws.com/images/png-transparent-default-avatar-thumbnail.png', '123 LLC',
        '789 Oak St, Anytown, USA', '2022-01-03 00:00:00', null, null, false, false, true) ON CONFLICT DO NOTHING;
END;