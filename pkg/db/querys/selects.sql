-- name: GetAllRunNames :many
SELECT id, name FROM runs;

-- name: GetAllRuns :many
SELECT id, name, game, category, attempts, active_split FROM runs;

-- name: GetSplitsByRunName :many
SELECT
  s.id,
  s.run_id,
  s.name,
  s.hit_count,
  s.pb_hit_count,
  s.diff,
  s.idx,
  s.save_file,
  (s.idx = r.active_split) AS is_active
FROM
splits AS s
JOIN
runs AS r
ON r.id = s.run_id
WHERE
r.name = ?
ORDER BY
s.idx;

-- name: GetSplitsByRunID :many
SELECT
  s.id,
  s.run_id,
  s.name,
  s.hit_count,
  s.pb_hit_count,
  s.diff,
  s.idx,
  s.save_file,
  (s.idx = r.active_split) AS is_active
FROM
splits AS s
JOIN
runs AS r
ON r.id = s.run_id
WHERE
s.run_id = ?
ORDER BY
s.idx;

-- name: GetRunByName :one
SELECT id, name, game, category, attempts, active_split FROM runs WHERE name = ?;

-- name: GetRunByID :one
SELECT id, name, game, category, attempts, active_split FROM runs WHERE id = ?;

-- name: GetSplitByRunIDAndIdx :one
SELECT id, run_id, name, hit_count, pb_hit_count, diff, idx, save_file 
FROM splits 
WHERE run_id = ? AND idx = ?;

-- name: GetSplitByRunNameAndIdx :one
SELECT s.id,
  s.run_id,
  s.name,
  s.hit_count,
  s.pb_hit_count,
  s.diff,
  s.idx,
  s.save_file
FROM splits s
JOIN runs r ON s.run_id = r.id
WHERE r.name = ? AND s.idx = ?;

-- name: GetActiveRun :one
SELECT id, name, game, category, attempts, active_split
FROM runs r JOIN misc_info m ON m.active_run = r.id;

-- name: GetRunAttemptsByName :one
SELECT attempts FROM runs WHERE name = ?;

-- name: GetRunIDByName :one
SELECT id FROM runs WHERE name = ?;

-- name: GetSplitByID :one
SELECT id, run_id, name, hit_count, pb_hit_count, diff, idx, save_file
FROM splits 
WHERE id = ?;

-- name: GetSplitTotalsByRunID :one
SELECT 
  SUM(hit_count) as total_hits,
  SUM(pb_hit_count) as total_pb,
  SUM(diff) as total_diff
FROM splits 
WHERE run_id = ?;
