BEGIN;
ALTER TABLE help_event ADD COLUMN is_deleted BOOLEAN DEFAULT false NOT NULL;
ALTER TABLE member_search ADD COLUMN event_type event NOT NULL;
END;