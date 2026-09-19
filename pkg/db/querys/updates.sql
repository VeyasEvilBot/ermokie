-- name: UpdateActiveRun :exec
UPDATE misc_info
SET active_run = (
  SELECT id FROM runs WHERE name = ?
);

-- name: UpdateActiveRunByID :exec
UPDATE misc_info
SET active_run = ?;

-- name: UpdateActiveSplitByID :exec
UPDATE runs
SET active_split = (
  SELECT idx FROM splits WHERE splits.id = ?
)
WHERE runs.id = ?;

-- name: AdvanceSplitInActiveRun :exec
UPDATE runs
SET active_split = MIN(
  active_split + 1,
  COALESCE((SELECT MAX(idx) FROM splits WHERE run_id = runs.id), active_split)
)
WHERE id = (SELECT active_run FROM misc_info WHERE one = 1);

-- name: GoBackSplitInActiveRun :exec
UPDATE runs
SET active_split = MAX(active_split - 1, 0)
WHERE id = (SELECT active_run FROM misc_info WHERE one = 1);

-- name: IncrementActiveSplitHit :exec
UPDATE splits
SET hit_count = COALESCE(hit_count, 0) + 1
WHERE run_id = (SELECT active_run FROM misc_info WHERE one = 1)
  AND idx = (SELECT active_split FROM runs WHERE runs.id = splits.run_id);

-- name: DecrementActiveSplitHit :exec
UPDATE splits
SET hit_count = MAX(COALESCE(hit_count, 0) - 1, 0)
WHERE run_id = (SELECT active_run FROM misc_info WHERE one = 1)
  AND idx = (SELECT active_split FROM runs WHERE runs.id = splits.run_id);

-- name: ResetActiveRunProgress :exec
UPDATE runs
SET active_split = 0,
    attempts = COALESCE(attempts, 0) + 1
WHERE id = (SELECT active_run FROM misc_info WHERE one = 1);

-- name: ResetActiveRunHits :exec
UPDATE splits
SET hit_count = 0
WHERE run_id = (SELECT active_run FROM misc_info WHERE one = 1);
