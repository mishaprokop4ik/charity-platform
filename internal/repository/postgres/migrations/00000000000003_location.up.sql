BEGIN;
    CREATE TABLE IF NOT EXISTS location (
      id bigserial PRIMARY KEY,
      country varchar default 'Ukraine' NOT NULL,
      area varchar,
      city varchar,
      district varchar,
      street varchar,
      home varchar,
      event_type event,
      event_id bigint
    );
END;