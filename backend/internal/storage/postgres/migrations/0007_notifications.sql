CREATE TABLE notifications (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type INT NOT NULL,
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, notification_type)
);
