-- 0008_backfill_notifications.sql

INSERT INTO notifications (user_id, notification_type, count)
SELECT receiver_id, 1, COUNT(*)
FROM invitations
GROUP BY receiver_id
ON CONFLICT (user_id, notification_type) 
DO UPDATE SET count = EXCLUDED.count;
