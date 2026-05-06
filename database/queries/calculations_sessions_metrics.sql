-- name: AggregateSessions :exec
INSERT INTO calculation_session_metrics (
    session_id,
    engine_version,
    calculation_runs,
    strategy_switches,
    input_modifications,
    pdf_requested,
    email_submitted,
    first_seen_at,
    last_seen_at
)
SELECT
    cs.session_id,
    cs.engine_version,

    COUNT(*) FILTER (WHERE ce.event_type = 'calculator_started') AS calculation_runs,
    COUNT(*) FILTER (WHERE ce.event_type = 'strategy_changed') AS strategy_switches,
    COUNT(*) FILTER (WHERE ce.event_type = 'input_modified_after_result') AS input_modifications,

    BOOL_OR(ce.event_type = 'pdf_requested') AS pdf_requested,
    BOOL_OR(ce.event_type = 'email_submitted') AS email_submitted,

    MIN(ce.created_at) AS first_seen_at,
    MAX(ce.created_at) AS last_seen_at

FROM calculation_sessions cs
JOIN calculation_events ce
  ON cs.session_id = ce.session_id

WHERE NOT EXISTS (
    SELECT 1 FROM calculation_session_metrics csm
    WHERE csm.session_id = cs.session_id
)

GROUP BY cs.session_id, cs.engine_version;

-- name: RepeatCalculationRate :one
SELECT
    COUNT(*) FILTER (WHERE calculation_runs >= 2)::float
    / NULLIF(COUNT(*), 0) AS repeat_rate
FROM calculation_session_metrics;

-- name: StrategySwitchRate :one
SELECT
    COUNT(*) FILTER (WHERE strategy_switches > 0)::float
    / NULLIF(COUNT(*), 0) AS switch_rate
FROM calculation_session_metrics;

-- name: PdfConversionRate :one
SELECT
    COUNT(*) FILTER (WHERE pdf_requested = TRUE)::float
    / NULLIF(COUNT(*), 0) AS pdf_rate
FROM calculation_session_metrics;

-- name: CommitmentRate :one
SELECT
    COUNT(*) FILTER (
        WHERE pdf_requested = TRUE AND email_submitted = TRUE
    )::float
    / NULLIF(COUNT(*), 0) AS commitment_rate
FROM calculation_session_metrics;

-- name: ReturningSessions :one
SELECT
    COUNT(*) FILTER (
        WHERE last_seen_at - first_seen_at > interval '1 day'
    )::float
    / NULLIF(COUNT(*), 0) AS return_rate
FROM calculation_session_metrics;

-- name: CleanupOldEvents :exec
DELETE FROM calculation_events
WHERE created_at < now() - interval '30 days'
  AND session_id IN (
      SELECT session_id FROM calculation_session_metrics
  );
