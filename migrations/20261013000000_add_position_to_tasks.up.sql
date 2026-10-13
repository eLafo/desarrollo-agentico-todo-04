ALTER TABLE tasks ADD COLUMN position integer NOT NULL DEFAULT 0;

UPDATE tasks
SET position = ordered.position
FROM (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC) AS position
  FROM tasks
) AS ordered
WHERE tasks.id = ordered.id;
