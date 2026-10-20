import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('añadir: el título no pasa de 120 caracteres', async ({ page }) => {
    // 1. La lista está vacía
    await expect(page.getByText('No hay tareas todavía.')).toBeVisible();

    // 2. Escribe de más: el campo se queda en 120 caracteres
    const titulo = 'a'.repeat(120);
    const campo = page.getByRole('textbox', { name: 'Nueva tarea' });
    await campo.fill(titulo + 'b'.repeat(10));
    await expect(campo).toHaveValue(titulo);

    // 3. Pulsa «Añadir»: se crea con los 120 caracteres
    await page.getByRole('button', { name: 'Añadir' }).click();
    await expect(page.getByRole('status')).toHaveText('Tarea creada.');
    await expect(page.locator('.task-title')).toHaveText([titulo]);

    // 4. Tras recargar sigue igual
    await page.reload();
    await expect(page.locator('.task-title')).toHaveText([titulo]);
  });
});
