CREATE VIEW v_secret_owner AS
SELECT
    'model' AS owner_kind,
    m.uuid AS owner_uuid,
    m.name AS owner_name,
    so.label,
    so.secret_id
FROM secret_model_owner AS so
JOIN model AS m -- this is a singleton
UNION ALL
SELECT
    'application' AS owner_kind,
    so.application_uuid AS owner_uuid,
    a.name AS owner_name,
    so.label,
    so.secret_id
FROM secret_application_owner AS so
JOIN application AS a ON so.application_uuid = a.uuid
UNION ALL
SELECT
    'unit' AS owner_kind,
    so.unit_uuid AS owner_uuid,
    u.name AS owner_name,
    so.label,
    so.secret_id
FROM secret_unit_owner AS so
JOIN unit AS u ON so.unit_uuid = u.uuid;

CREATE VIEW v_secret_metadata AS
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
    so.owner_uuid,
    so.owner_name,
    so.label
FROM secret_metadata AS sm
JOIN secret_rotate_policy AS rp ON sm.rotate_policy_id = rp.id
JOIN secret_revision AS sr ON sm.secret_id = sr.secret_id
LEFT JOIN secret_revision_expire AS sre ON sr.uuid = sre.revision_uuid
LEFT JOIN secret_rotation AS sro ON sm.secret_id = sro.secret_id
LEFT JOIN v_secret_owner AS so ON sm.secret_id = so.secret_id;
