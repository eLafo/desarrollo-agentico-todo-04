import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('eliminar pide confirmación', async ({ page, tasks }) => {
    await tasks([{ title: 'Enviar la factura' }, { title: 'Revisar el presupuesto' }]);
    await page.reload();
    const eliminar = page.getByRole('button', { name: 'Eliminar tarea: Enviar la factura' });

    // 1. Pulsa eliminar y cancela: la tarea sigue
    page.once('dialog', (dialog) => dialog.dismiss());
    await eliminar.click();
    await expect(page.locator('.task-title')).toHaveText(['Enviar la factura', 'Revisar el presupuesto']);

    // 2. Pulsa eliminar y acepta: la tarea desaparece
    page.once('dialog', async (dialog) => {
      expect(dialog.message()).toBe('¿Seguro que quieres eliminar esta tarea?');
      await dialog.accept();
    });
    await eliminar.click();
    await expect(page.getByRole('status')).toHaveText('Tarea eliminada: Enviar la factura.');
    await expect(page.locator('.task-title')).toHaveText(['Revisar el presupuesto']);
  });
});
