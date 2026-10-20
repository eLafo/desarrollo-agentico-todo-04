import { test as base, expect } from '@playwright/test';
import pg from 'pg';

export { expect };

type Task = { title: string; completed?: boolean };

type Fixtures = {
  // Deja la tabla con exactamente estas tareas, en este orden de creación.
  tasks: (tasks: Task[]) => Promise<void>;
};

// La misma base de datos que usa la aplicación con GO_ENV=test (database.yml).
function client() {
  return new pg.Client({
    host: process.env.DB_HOST ?? '127.0.0.1',
    port: Number(process.env.DB_PORT ?? 5432),
    user: 'postgres',
    password: 'postgres',
    database: 'todo_test',
  });
}

// Cada test empieza con la lista vacía y la página abierta. Si necesita tareas, las crea con
// `tasks` y recarga la página.
export const test = base.extend<Fixtures>({
  tasks: async ({}, use) => {
    const db = client();
    await db.connect();
    await db.query('DELETE FROM tasks');
    await use(async (tasks) => {
      await db.query('DELETE FROM tasks');
      for (const [i, task] of tasks.entries()) {
        await db.query(
          `INSERT INTO tasks (id, title, completed, created_at, updated_at)
           VALUES (gen_random_uuid(), $1, $2, now() + $3 * interval '1 millisecond', now())`,
          [task.title, task.completed ?? false, i],
        );
      }
    });
    await db.end();
  },
  page: async ({ page, tasks }, use) => {
    void tasks;
    await page.goto('/');
    await use(page);
  },
});
