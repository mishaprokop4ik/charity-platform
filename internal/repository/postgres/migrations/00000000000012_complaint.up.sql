BEGIN;
ALTER TABLE help_event ADD COLUMN is_banned BOOLEAN DEFAULT false NOT NULL;
ALTER TABLE propositional_event ADD COLUMN is_banned BOOLEAN DEFAULT false NOT NULL;
CREATE TABLE IF NOT EXISTS complaints (
    id bigserial PRIMARY KEY,
    description varchar,
    event_type event,
    created_by bigint,
    event_id bigint,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT author_fk FOREIGN KEY(created_by) REFERENCES members(id)
);
END;