INSERT INTO members (full_name, telephone, telegram_username, email, password, image_path, company_name, address, created_at, updated_at, deleted_at, is_deleted, is_admin, is_activated)
VALUES
    ('John Doe', '+380671234567', '@john_doe', 'john.doe@example.com', 'password123', '/images/john_doe.png', 'ABC Company', '123 Main St, Anytown, USA', '2022-01-01 00:00:00', null, null, false, true, true),
    ('Jane Smith', '+380671234567', '@jane_smith', 'jane.smith@example.com', 'password456', '/images/jane_smith.png', 'XYZ Corporation', '456 Elm St, Anytown, USA', '2022-01-02 00:00:00', null, null, false, false, true),
    ('Bob Johnson', '+380671234567', '@bob_johnson', 'bob.johnson@example.com', 'password789', '/images/bob_johnson.png', '123 LLC', '789 Oak St, Anytown, USA', '2022-01-03 00:00:00', null, null, false, false, true);

INSERT INTO propositional_event (title, description, creation_date, competition_date, status, max_concurrent_requests, remaining_helps, author_id, category, is_deleted)
VALUES
    ('Fundraising Event', 'A charity event to raise funds for a local non-profit organization.', '2022-02-01 00:00:00', '2022-03-01 00:00:00', 'active', 50, 10, 1, 'charity', false),
    ('Volunteer Day', 'A day of volunteering at a local community center.', '2022-03-01 00:00:00', '2022-03-01 00:00:00', 'inactive', 10, 0, 2, 'volunteering', false),
    ('Community Cleanup', 'A day of cleaning up a local park.', '2022-04-01 00:00:00', '2022-04-01 00:00:00', 'active', 20, 5, 3, 'environmental', false);

INSERT INTO help_event (title, description, creation_date, author_id, category, address, competition_date, rate)
VALUES
    ('Emergency Assistance', 'Assistance needed for an emergency situation.', '2022-02-01 00:00:00', 1, 'emergency', '123 Main St, Anytown, USA', '2022-02-05 00:00:00', 'normal'),
    ('Moving Help', 'Assistance needed for moving to a new home.', '2022-03-01 00:00:00', 2, 'moving', '456 Elm St, Anytown, USA', '2022-03-05 00:00:00', 'fast'),
    ('Yard Work', 'Assistance needed for yard work.', '2022-04-01 00:00:00', 3, 'gardening', '789 Oak St, Anytown, USA', '2022-04-05 00:00:00', 'very fast');

INSERT INTO tag (title, event_id, event_type)
VALUES
    ('marathon', 1, 'proposal-event'),
    ('charity', 1, 'proposal-event'),
    ('blood donation', 2, 'proposal-event'),
    ('art', 3, 'proposal-event'),
    ('homelessness', 4, 'help'),
    ('beach', 5, 'help'),
    ('education', 6, 'help');

INSERT INTO tag_value (tag_id, value)
VALUES
    (1, 'Sports'),
    (2, 'Charity'),
    (3, 'Health'),
    (4, 'Arts'),
    (5, 'Community'),
    (6, 'Environment'),
    (7, 'Education');