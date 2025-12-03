-- name: CreateMeeting :one
INSERT INTO meetings (title, raw_notes, source, metadata)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at;

-- name: GetMeeting :one
SELECT * FROM meetings WHERE id = $1;

-- name: UpdateMeetingStatus :exec
UPDATE meetings
SET 
  processing_status = $2,
  processing_error = CASE 
    WHEN $3::text = '' THEN NULL 
    ELSE $3 
  END,
  processed_at = CASE 
    WHEN $2 = 'done' OR $2 = 'failed' THEN NOW()
    ELSE processed_at 
  END
WHERE id = $1;

-- name: CreateTheme :one
INSERT INTO themes (name) VALUES ($1) RETURNING id, name, created_at;

-- name: GetThemeByName :one
SELECT id, name, created_at FROM themes WHERE name = $1;

-- name: CreateOpportunity :one
INSERT INTO opportunities (
    user_segment, struggle, why_it_matters, workaround, theme_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOpportunity :one
SELECT 
    o.*,
    
    t.name AS theme_name,
    COUNT(oe.id) AS evidence_count
FROM opportunities o
LEFT JOIN themes t ON o.theme_id = t.id
LEFT JOIN opportunity_evidence oe ON oe.opportunity_id = o.id
WHERE o.id = $1
GROUP BY o.id, t.name;

-- name: ListEvidenceByOpportunity :many
SELECT 
    oe.id,
    oe.quote,
    oe.context,
    oe.created_at,
    
    m.id            AS meeting_id,
    m.title         AS meeting_title,
    m.created_at    AS meeting_date,
    
    COUNT(*) OVER() AS total_count

FROM opportunity_evidence oe
JOIN meetings m ON oe.meeting_id = m.id
WHERE oe.opportunity_id = $1
ORDER BY oe.created_at DESC;

-- name: AddEvidence :exec
INSERT INTO opportunity_evidence (
    opportunity_id, meeting_id, quote, context
) VALUES ($1, $2, $3, $4);

-- name: ListAllOpportunitiesForDeduplication :many
SELECT 
    o.id::text AS opportunity_id,
    o.struggle AS struggle,
    COALESCE(t.name, 'No theme') AS theme_name
FROM opportunities o
LEFT JOIN themes t ON o.theme_id = t.id
ORDER BY o.created_at DESC;

-- name: ListRecentOpportunities :many
SELECT 
    o.*,
    t.name AS theme_name,
    COALESCE(ev_count.cnt, 0) AS evidence_count
FROM opportunities o

LEFT JOIN themes t ON o.theme_id = t.id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS cnt
    FROM opportunity_evidence oe
    WHERE oe.opportunity_id = o.id
) ev_count ON true

WHERE o.created_at >= NOW() - INTERVAL '24 hours'
ORDER BY o.created_at DESC;

-- name: ListTopThemesThisWeek :many
SELECT 
    t.name AS theme_name,
    COUNT(*) AS opportunity_count,
    -- for better ranking
    COUNT(*) + COUNT(*) FILTER (WHERE o.created_at >= NOW() - INTERVAL '3 days') AS score
FROM opportunities o
JOIN themes t ON o.theme_id = t.id
WHERE o.created_at >= date_trunc('week', CURRENT_DATE)
GROUP BY t.id, t.name
HAVING COUNT(*) >= 1
ORDER BY score DESC, opportunity_count DESC, t.name
LIMIT 10;

-- name: ListTopOpportunitiesByTheme :many
SELECT 
    o.id,
    o.user_segment,
    o.struggle,
    o.why_it_matters,
    o.workaround,
    o.created_at,
    COUNT(oe.id) AS evidence_count
FROM opportunities o
LEFT JOIN opportunity_evidence oe ON oe.opportunity_id = o.id
JOIN themes t ON o.theme_id = t.id
WHERE LOWER(t.name) = LOWER($1)
GROUP BY o.id, o.struggle, o.created_at
ORDER BY evidence_count DESC, o.created_at DESC
LIMIT $2;
