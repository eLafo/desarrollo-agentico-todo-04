import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('añadir un título que sigue pasando del límite con los espacios juntos', async ({ page }) => {
    // 201 caracteres con un solo espacio entre palabras; 205 tal como se escribe
    const escrito = 'Comprar     pan ' + 'x'.repeat(189);

    // 1. Escribe el título tecla a tecla y pulsa «Añadir»
    await page.getByRole('textbox', { name: 'Nueva tarea' }).pressSequentially(escrito);
    await page.getByRole('button', { name: 'Añadir' }).click();

    // 2. Se rechaza con el mensaje de siempre y no se crea ninguna tarea
    await expect(page.getByRole('alert')).toHaveText('El título no puede superar los 200 caracteres.');
    await expect(page.locator('.task-title')).toHaveCount(0);
  });
});
