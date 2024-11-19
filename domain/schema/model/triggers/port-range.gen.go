// Code generated by triggergen. DO NOT EDIT.

package triggers

import (
	"fmt"

	"github.com/juju/juju/core/database/schema"
)


// ChangeLogTriggersForPortRange generates the triggers for the
// port_range table.
func ChangeLogTriggersForPortRange(columnName string, namespaceID int) func() schema.Patch {
	return func() schema.Patch {
		return schema.MakePatch(fmt.Sprintf(`
-- insert namespace for PortRange
INSERT INTO change_log_namespace VALUES (%[2]d, 'port_range', 'PortRange changes based on %[1]s');

-- insert trigger for PortRange
CREATE TRIGGER trg_log_port_range_insert
AFTER INSERT ON port_range FOR EACH ROW
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (1, %[2]d, NEW.%[1]s, DATETIME('now'));
END;

-- update trigger for PortRange
CREATE TRIGGER trg_log_port_range_update
AFTER UPDATE ON port_range FOR EACH ROW
WHEN 
	(NEW.uuid != OLD.uuid OR (NEW.uuid IS NOT NULL AND OLD.uuid IS NULL) OR (NEW.uuid IS NULL AND OLD.uuid IS NOT NULL)) OR
	NEW.unit_endpoint_uuid != OLD.unit_endpoint_uuid OR
	NEW.protocol_id != OLD.protocol_id OR
	(NEW.from_port != OLD.from_port OR (NEW.from_port IS NOT NULL AND OLD.from_port IS NULL) OR (NEW.from_port IS NULL AND OLD.from_port IS NOT NULL)) OR
	(NEW.to_port != OLD.to_port OR (NEW.to_port IS NOT NULL AND OLD.to_port IS NULL) OR (NEW.to_port IS NULL AND OLD.to_port IS NOT NULL)) 
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (2, %[2]d, OLD.%[1]s, DATETIME('now'));
END;
-- delete trigger for PortRange
CREATE TRIGGER trg_log_port_range_delete
AFTER DELETE ON port_range FOR EACH ROW
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (4, %[2]d, OLD.%[1]s, DATETIME('now'));
END;`, columnName, namespaceID))
	}
}
