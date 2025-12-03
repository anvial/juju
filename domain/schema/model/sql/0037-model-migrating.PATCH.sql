CREATE TABLE model_migrating (
    uuid TEXT NOT NULL PRIMARY KEY,
    model_uuid TEXT NOT NULL,
    CONSTRAINT fk_model_migrating_model
    FOREIGN KEY (model_uuid)
    REFERENCES model (uuid)
);

CREATE UNIQUE INDEX idx_model_migrating_model_uuid
ON model_migrating (model_uuid);
