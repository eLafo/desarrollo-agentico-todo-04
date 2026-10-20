import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('añadir una tarea', async ({ page }) => {
    // 1. La lista está vacía
    await expect(page.getByText('No hay tareas todavía.')).toBeVisible();

    // 2. Escribe el título y pulsa «Añadir»
    await page.getByRole('textbox', { name: 'Nueva tarea' }).fill('Comprar pan');
    await page.getByRole('button', { name: 'Añadir' }).click();

    // 3. La tarea aparece pendiente, con el aviso y el contador
    await expect(page.getByRole('status')).toHaveText('Tarea creada.');
    await expect(page.locator('.task-title')).toHaveText(['Comprar pan']);
    await expect(page.getByRole('button', { name: 'Completar tarea: Comprar pan' })).toBeVisible();
    await expect(page.locator('.list-count')).toHaveText('1 pendiente de 1 tarea');
  });
});
