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
