BEGIN;
ALTER TABLE help_event ADD COLUMN end_date timestamp;
ALTER TABLE propositional_event ADD COLUMN end_date timestamp;
END;