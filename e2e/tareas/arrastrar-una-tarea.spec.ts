import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('arrastrar una tarea', async ({ page, tasks }) => {
    await tasks([{ title: 'Enviar la factura' }, { title: 'Revisar el presupuesto' }, { title: 'Llamar al proveedor' }]);
    await page.reload();
    const titulos = page.locator('.task-title');
    const llamar = page.getByRole('listitem').filter({ hasText: 'Llamar al proveedor' });
    const factura = page.getByRole('listitem').filter({ hasText: 'Enviar la factura' });

    // 1. Arrastra «Llamar al proveedor» a la mitad de arriba de «Enviar la factura»
    const guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await llamar.dragTo(factura, { targetPosition: { x: 100, y: 5 } });
    await guardado;

    // 2. La lista muestra el nuevo orden
    await expect(titulos).toHaveText(['Llamar al proveedor', 'Enviar la factura', 'Revisar el presupuesto']);

    // 3. Al recargar sigue el mismo orden
    await page.reload();
    await expect(titulos).toHaveText(['Llamar al proveedor', 'Enviar la factura', 'Revisar el presupuesto']);
  });
});
