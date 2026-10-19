import { test, expect } from '../fixtures';

// El título más largo que se puede guardar (200 caracteres), con < y &.
const largo = 'Revisar <factura> & presupuesto '.repeat(6) + 'de enero';

test.describe('Tareas', () => {
  test('el aviso de eliminar dice qué tarea', async ({ page, tasks }) => {
    await tasks([{ title: 'Comprar pan' }, { title: largo }, { title: 'Llamar al gestor' }]);
    await page.reload();
    page.on('dialog', (dialog) => dialog.accept());

    // 1. Elimina «Comprar pan»: el aviso dice cuál
    await page.getByRole('button', { name: 'Eliminar tarea: Comprar pan' }).click();
    await expect(page.getByRole('status')).toHaveText('Tarea eliminada: Comprar pan.');
    await expect(page.locator('.task-title')).toHaveText([largo, 'Llamar al gestor']);

    // 2. Elimina la del título largo: el aviso lo muestra entero, con < y & como texto
    await page.getByRole('button', { name: `Eliminar tarea: ${largo}` }).click();
    await expect(page.getByRole('status')).toHaveText(`Tarea eliminada: ${largo}.`);
    await expect(page.locator('.task-title')).toHaveText(['Llamar al gestor']);
  });
});
