import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('editar un título que solo pasa del límite por los espacios repetidos', async ({ page, tasks }) => {
    await tasks([{ title: 'Llamar al proveedor' }]);
    await page.reload();
    // 200 caracteres con un solo espacio entre palabras; 206 tal como se escribe
    const juntado = 'Llamar al proveedor ' + 'x'.repeat(180);
    const escrito = 'Llamar   al   proveedor   ' + 'x'.repeat(180);

    // 1. Abre la edición y escribe el título nuevo tecla a tecla: el campo no lo corta
    await page.getByRole('link', { name: 'Editar tarea: Llamar al proveedor' }).click();
    const campo = page.getByRole('textbox', { name: 'Título' });
    await campo.clear();
    await campo.pressSequentially(escrito);
    await expect(campo).toHaveValue(escrito);

    // 2. Guarda: la tarea se actualiza
    await page.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByRole('status')).toHaveText('Tarea actualizada.');
    await expect(page.locator('.task-title')).toHaveText([juntado]);

    // 3. Al volver a editarla, el título guardado tiene un solo espacio entre palabras
    await page.getByRole('link', { name: `Editar tarea: ${juntado}` }).click();
    await expect(campo).toHaveValue(juntado);
  });
});
