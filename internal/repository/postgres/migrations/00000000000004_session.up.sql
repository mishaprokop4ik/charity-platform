BEGIN;
    CREATE TABLE IF NOT EXISTS member_session (
        refresh_token varchar,
        member_id bigint,
        expires_at timestamp NOT NULL,
        PRIMARY KEY(member_id),
        CONSTRAINT members_fk FOREIGN KEY(member_id) REFERENCES members(id)
            ON DELETE CASCADE ON UPDATE CASCADE
    );
END;