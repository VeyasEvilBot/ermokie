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
SET active_split = active_split + 1
  WHERE (SELECT active_run FROM misc_info);

-- name: GoBackSplitInActiveRun :exec
UPDATE runs
SET active_split = active_split - 1
  WHERE (SELECT active_run FROM misc_info);
