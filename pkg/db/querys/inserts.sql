-- name: InsertRun :one
INSERT INTO runs (name, attempts, active_split, game, category) VALUES (?, ?, ?, ?, ?) RETURNING id;

-- name: InsertSplit :exec
INSERT INTO splits (run_id, name, hit_count, pb_hit_count, idx, save_file) VALUES (?, ?, ?, ?, ?, ?);
