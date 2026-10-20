import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('editar: el título no pasa de 120 caracteres', async ({ page, tasks }) => {
    await tasks([{ title: 'Llamar al proveedor' }]);
    await page.reload();

    // 1. Abre la edición y comprueba la ayuda con el límite
    await page.getByRole('link', { name: 'Editar tarea: Llamar al proveedor' }).click();
    await expect(page.getByRole('heading', { name: 'Editar tarea' })).toBeVisible();
    await expect(page.getByText('Hasta 120 caracteres. Los demás datos de la tarea no cambian.')).toBeVisible();

    // 2. Escribe de más: el campo se queda en 120 caracteres
    const titulo = 'a'.repeat(120);
    const campo = page.getByRole('textbox', { name: 'Título' });
    await campo.fill(titulo + 'b'.repeat(10));
    await expect(campo).toHaveValue(titulo);

    // 3. Guarda: la tarea queda con los 120 caracteres
    await page.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByRole('status')).toHaveText('Tarea actualizada.');
    await expect(page.locator('.task-title')).toHaveText([titulo]);

    // 4. Tras recargar sigue igual
    await page.reload();
    await expect(page.locator('.task-title')).toHaveText([titulo]);
  });
});
