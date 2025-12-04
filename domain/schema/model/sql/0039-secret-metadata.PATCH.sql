CREATE VIEW v_secret_metadata AS
WITH
secret_owner AS (
    SELECT 'model' AS owner_kind, (SELECT uuid FROM model LIMIT 1) AS owner_id, so.label, so.secret_id
    FROM   secret_model_owner AS so
    UNION ALL
    SELECT 'application' AS owner_kind, a.name AS owner_id, so.label, so.secret_id
    FROM   secret_application_owner AS so
    JOIN   application AS a ON so.application_uuid = a.uuid
    UNION ALL
    SELECT 'unit' AS owner_kind, u.name AS owner_id, so.label, so.secret_id
    FROM   secret_unit_owner AS so
    JOIN   unit AS u ON so.unit_uuid = u.uuid
)
SELECT
    sm.secret_id,
    sm.version,
    sm.description,
    sm.auto_prune,
    sm.latest_revision_checksum,
    sm.create_time,
    sm.update_time,
    rp.policy,
    sro.next_rotation_time,
    sre.expire_time,
    sr.revision,
    so.owner_kind,
    so.owner_id,
    so.label
FROM        secret_metadata AS sm
JOIN        secret_revision AS sr ON sm.secret_id = sr.secret_id
LEFT JOIN   secret_revision_expire AS sre ON sre.revision_uuid = sr.uuid
LEFT JOIN   secret_rotate_policy AS rp ON rp.id = sm.rotate_policy_id
LEFT JOIN   secret_rotation AS sro ON sro.secret_id = sm.secret_id
LEFT JOIN   secret_owner AS so ON so.secret_id = sm.secret_id;