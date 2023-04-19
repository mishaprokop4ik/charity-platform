-- comments
CREATE OR REPLACE FUNCTION check_event_exists()
    RETURNS TRIGGER AS $$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS (SELECT 1 FROM proposal_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS (SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_event_exists_trigger
    BEFORE INSERT OR UPDATE ON comment
    FOR EACH ROW EXECUTE FUNCTION check_event_exists();
-- transactions
CREATE OR REPLACE FUNCTION transaction_check_event_exists()
    RETURNS TRIGGER AS $$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS (SELECT 1 FROM proposal_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS (SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transaction_event_exists_trigger
    BEFORE INSERT OR UPDATE ON transaction
    FOR EACH ROW EXECUTE FUNCTION transaction_check_event_exists();
-- notification
CREATE OR REPLACE FUNCTION notification_check_event_exists()
    RETURNS TRIGGER AS $$
BEGIN
    -- Check if the event with corresponding id exists based on event_type
    IF NEW.event_type = 'proposal-event' THEN
        IF NOT EXISTS (SELECT 1 FROM proposal_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Proposal event with id % does not exist', NEW.event_id;
        END IF;
    ELSIF NEW.event_type = 'help' THEN
        IF NOT EXISTS (SELECT 1 FROM help_event WHERE id = NEW.event_id) THEN
            RAISE EXCEPTION 'Help event with id % does not exist', NEW.event_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid event_type: %', NEW.event_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notification_event_exists_trigger
    BEFORE INSERT OR UPDATE ON notification
    FOR EACH ROW EXECUTE FUNCTION notification_check_event_exists();