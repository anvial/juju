CREATE VIEW v_model_config AS
SELECT
    mc."key",
    mc.value
FROM model_config AS mc
UNION
SELECT
    'agent-stream' AS "key",
    mas.name AS value
FROM agent_version AS mv
JOIN agent_stream AS mas ON mv.stream_id = mas.id
UNION
SELECT
    'agent-version' AS "key",
    mv.target_version AS value
FROM agent_version AS mv;

-- Remove the trigger because it wasn't looking for changes to stream_id.
DROP TRIGGER trg_log_agent_version_update;

CREATE TRIGGER trg_log_agent_version_update
AFTER UPDATE ON agent_version FOR EACH ROW
WHEN
	NEW.stream_id != OLD.stream_id OR
    NEW.target_version != OLD.target_version
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (2, 10027, NEW.target_version, DATETIME('now', 'utc'));
END;
