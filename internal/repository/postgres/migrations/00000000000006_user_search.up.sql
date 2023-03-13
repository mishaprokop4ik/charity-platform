BEGIN;
    CREATE TABLE IF NOT EXISTS member_search (
        id bigserial PRIMARY KEY,
        title varchar(255),
        member_id bigint,
        CONSTRAINT member_id FOREIGN KEY(member_id) REFERENCES members(id)
            ON DELETE CASCADE ON UPDATE CASCADE
    );
    CREATE TABLE IF NOT EXISTS member_search_value (
        id bigserial PRIMARY KEY,
        member_search_id bigint,
        value varchar(255),
        CONSTRAINT member_search_id FOREIGN KEY(member_search_id) REFERENCES member_search(id)
            ON DELETE CASCADE ON UPDATE CASCADE
    );
END;