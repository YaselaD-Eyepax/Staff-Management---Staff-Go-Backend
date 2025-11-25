INSERT INTO event_tags (event_id, tag)
VALUES 
    (NULL, 'general'),
    (NULL, 'announcement'),
    (NULL, 'reminder'),
    (NULL, 'deadline'),
    (NULL, 'academic'),
    (NULL, 'holiday'),
    (NULL, 'emergency')
ON CONFLICT DO NOTHING;