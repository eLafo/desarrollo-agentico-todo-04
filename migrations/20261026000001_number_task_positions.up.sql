UPDATE tasks
SET position = ordered.n - 1
FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY created_at, id) AS n FROM tasks) AS ordered
WHERE tasks.id = ordered.id;
