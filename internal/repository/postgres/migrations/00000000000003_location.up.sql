BEGIN;
    CREATE TABLE IF NOT EXISTS location (
      id bigserial PRIMARY KEY,
      country varchar default 'Ukraine' NOT NULL,
      area varchar NOT NULL,
      city varchar NOT NULL,
      district varchar NOT NULL,
      street varchar NOT NULL,
      home varchar NOT NULL,
      event_type event,
      event_id bigint
    );
END;