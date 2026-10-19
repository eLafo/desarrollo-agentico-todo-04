import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('añadir un título que solo pasa del límite por los espacios repetidos', async ({ page }) => {
    // 200 caracteres con un solo espacio entre palabras; 204 tal como se escribe
    const juntado = 'Comprar pan ' + 'x'.repeat(188);
    const escrito = 'Comprar     pan ' + 'x'.repeat(188);
    const campo = page.getByRole('textbox', { name: 'Nueva tarea' });

    // 1. Escribe el título tecla a tecla: el campo no lo corta
    await campo.pressSequentially(escrito);
    await expect(campo).toHaveValue(escrito);

    // 2. Pulsa «Añadir»: la tarea se crea
    await page.getByRole('button', { name: 'Añadir' }).click();
    await expect(page.getByRole('status')).toHaveText('Tarea creada.');
    await expect(page.locator('.task-title')).toHaveText([juntado]);

    // 3. Al editarla, el título guardado tiene un solo espacio entre palabras
    await page.getByRole('link', { name: `Editar tarea: ${juntado}` }).click();
    await expect(page.getByRole('textbox', { name: 'Título' })).toHaveValue(juntado);
  });
});
