// Code generated by triggergen. DO NOT EDIT.

package triggers

import (
	"fmt"

	"github.com/juju/juju/core/database/schema"
)


// ChangeLogTriggersForMachineCloudInstance generates the triggers for the
// machine_cloud_instance table.
func ChangeLogTriggersForMachineCloudInstance(columnName string, namespaceID int) func() schema.Patch {
	return func() schema.Patch {
		return schema.MakePatch(fmt.Sprintf(`
-- insert trigger for MachineCloudInstance
CREATE TRIGGER trg_log_machine_cloud_instance_insert
AFTER INSERT ON machine_cloud_instance FOR EACH ROW
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (1, %[2]d, NEW.%[1]s, DATETIME('now'));
END;

-- update trigger for MachineCloudInstance
CREATE TRIGGER trg_log_machine_cloud_instance_update
AFTER UPDATE ON machine_cloud_instance FOR EACH ROW
WHEN 
	NEW.instance_id != OLD.instance_id OR
	(NEW.arch != OLD.arch OR (NEW.arch IS NOT NULL AND OLD.arch IS NULL) OR (NEW.arch IS NULL AND OLD.arch IS NOT NULL)) OR
	(NEW.mem != OLD.mem OR (NEW.mem IS NOT NULL AND OLD.mem IS NULL) OR (NEW.mem IS NULL AND OLD.mem IS NOT NULL)) OR
	(NEW.root_disk != OLD.root_disk OR (NEW.root_disk IS NOT NULL AND OLD.root_disk IS NULL) OR (NEW.root_disk IS NULL AND OLD.root_disk IS NOT NULL)) OR
	(NEW.root_disk_source != OLD.root_disk_source OR (NEW.root_disk_source IS NOT NULL AND OLD.root_disk_source IS NULL) OR (NEW.root_disk_source IS NULL AND OLD.root_disk_source IS NOT NULL)) OR
	(NEW.cpu_cores != OLD.cpu_cores OR (NEW.cpu_cores IS NOT NULL AND OLD.cpu_cores IS NULL) OR (NEW.cpu_cores IS NULL AND OLD.cpu_cores IS NOT NULL)) OR
	(NEW.cpu_power != OLD.cpu_power OR (NEW.cpu_power IS NOT NULL AND OLD.cpu_power IS NULL) OR (NEW.cpu_power IS NULL AND OLD.cpu_power IS NOT NULL)) OR
	(NEW.availability_zone_uuid != OLD.availability_zone_uuid OR (NEW.availability_zone_uuid IS NOT NULL AND OLD.availability_zone_uuid IS NULL) OR (NEW.availability_zone_uuid IS NULL AND OLD.availability_zone_uuid IS NOT NULL)) OR
	(NEW.virt_type != OLD.virt_type OR (NEW.virt_type IS NOT NULL AND OLD.virt_type IS NULL) OR (NEW.virt_type IS NULL AND OLD.virt_type IS NOT NULL)) 
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (2, %[2]d, OLD.%[1]s, DATETIME('now'));
END;

-- delete trigger for MachineCloudInstance
CREATE TRIGGER trg_log_machine_cloud_instance_delete
AFTER DELETE ON machine_cloud_instance FOR EACH ROW
BEGIN
    INSERT INTO change_log (edit_type_id, namespace_id, changed, created_at)
    VALUES (4, %[2]d, OLD.%[1]s, DATETIME('now'));
END;`, columnName, namespaceID))
	}
}
