BEGIN;
    CREATE TABLE IF NOT EXISTS notification (
                                                id BIGSERIAL PRIMARY KEY,
                                                event_type event NOT NULL,
                                                event_id BIGINT NOT NULL,
                                                action varchar NOT NULL,
                                                transaction_id INTEGER NOT NULL,
                                                new_status varchar NOT NULL,
                                                is_read BOOLEAN DEFAULT false NOT NULL,
                                                creation_time TIMESTAMP DEFAULT now() NOT NULL,
                                                member_id BIGINT NOT NULL
    );
END;