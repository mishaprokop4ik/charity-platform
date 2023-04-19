BEGIN;
    CREATE type need_unit AS ENUM ('kilogram', 'liter', 'item', 'work');

    CREATE TABLE IF NOT EXISTS need (
        id bigserial PRIMARY KEY,
        title varchar NOT NULL,
        amount integer,
        unit need_unit NOT NULL,
        help_event_id bigint,
        received integer,
        received_total integer,
        transaction_id bigint,
        CONSTRAINT event_fk FOREIGN KEY(help_event_id) REFERENCES help_event(id)
            ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT transaction_fk FOREIGN KEY(transaction_id) REFERENCES transaction(id)
            ON DELETE CASCADE ON UPDATE CASCADE
    );

    ALTER TABLE help_event RENAME COLUMN competition_date TO completion_time;
    ALTER TABLE help_event RENAME COLUMN creation_date TO created_at;
    ALTER TABLE help_event RENAME COLUMN author_id TO created_by;
    ALTER TABLE help_event ADD COLUMN status event_status DEFAULT 'inactive';
    ALTER TABLE help_event ADD COLUMN image_path varchar;
    ALTER TABLE help_event DROP COLUMN rate;
ALTER TABLE help_event DROP COLUMN category;
END;