--- cleanup

-- name: DeleteOldMessages :exec
DELETE FROM message_queue
WHERE consumed_by IS NOT NULL
AND created_at < datetime('now', ?);

-- name: DeleteOldUnconsumedMessages :exec
DELETE FROM message_queue
WHERE consumed_by IS NULL
AND created_at < datetime('now', ?);

-- name: DeleteOldSubscriptions :exec
DELETE FROM subscriptions
WHERE timestamp < datetime('now', ?);

-- name: DeleteSubscriptionsByTopic :exec
DELETE FROM subscriptions
WHERE topic = ?;

-- name: DeleteOldMessageLogs :exec
DELETE FROM message_log
WHERE created_at < datetime('now', ?);
