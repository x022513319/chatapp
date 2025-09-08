ALTER TABLE rooms
ADD CONSTRAINT rooms_workspace_name_unique
UNIQUE (workspace_id, name);