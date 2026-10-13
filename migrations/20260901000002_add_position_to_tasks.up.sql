ALTER TABLE tasks ADD COLUMN position integer NOT NULL DEFAULT 0;

UPDATE tasks
SET position = ranked.rn
FROM (
  SELECT id, row_number() OVER (ORDER BY created_at ASC, id ASC) AS rn
  FROM tasks
) AS ranked
WHERE tasks.id = ranked.id;

CREATE INDEX tasks_position_idx ON tasks (position);
