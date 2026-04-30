-- name: CreateAlert :one
INSERT INTO alerts (user_id,symbol,target_price,direction)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: ListAlertsByUser :many
SELECT * FROM alerts WHERE user_id = $1;

-- name: DeleteAlert :exec
DELETE FROM alerts WHERE id = $1;

-- name: GetActiveAlertsBySymbol :many
SELECT * FROM alerts WHERE symbol = $1 AND active = true;

-- name: CreateTriggeredAlert :one
INSERT INTO triggered_alerts (alert_id, triggered_price)
VALUES ($1, $2)
RETURNING *;

-- name: DeactivateAlert :exec
UPDATE alerts SET active = false WHERE id = $1;

-- name: GetAllActiveAlerts :many
SELECT * FROM alerts WHERE active = true;