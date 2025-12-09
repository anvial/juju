CREATE TABLE model_migration_import (
    uuid TEXT NOT NULL PRIMARY KEY,
    model_uuid TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_model_migration_import
ON model_migration_import (model_uuid);

DROP VIEW v_model_state;

-- v_model_state exists to provide a simple view over the states that are
-- needed to calculate a model's status.
CREATE VIEW v_model_state AS
SELECT
    m.uuid,
    cc.invalid AS cloud_credential_invalid,
    cc.invalid_reason AS cloud_credential_invalid_reason,
    IIF(mmi.model_uuid IS NOT NULL, TRUE, FALSE) AS migrating,
    IIF(l.id = 1, TRUE, FALSE) AS destroying
FROM model AS m
JOIN life AS l ON m.life_id = l.id
LEFT JOIN cloud_credential AS cc ON m.cloud_credential_uuid = cc.uuid
LEFT JOIN model_migration_import AS mmi ON m.uuid = mmi.model_uuid
WHERE m.activated = TRUE;
