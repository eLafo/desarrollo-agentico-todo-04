import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('las pendientes cambian al completar y reabrir', async ({ page, tasks }) => {
    await tasks([
      { title: 'Enviar la factura' },
      { title: 'Revisar el presupuesto' },
      { title: 'Pagar el alquiler', completed: true },
    ]);
    await page.reload();
    const recuento = page.locator('.list-count');
    const pagina = page.locator('body');
    await expect(recuento).toHaveText('2 pendientes de 3 tareas');

    // 1. Completa «Enviar la factura»: el texto baja sin recargar la página
    await page.evaluate(() => { document.body.dataset.sinRecargar = 'true'; });
    await page.getByRole('button', { name: 'Completar tarea: Enviar la factura' }).click();
    await expect(recuento).toHaveText('1 pendiente de 3 tareas');
    await expect(pagina).toHaveAttribute('data-sin-recargar', 'true');

    // 2. Al recargar, el texto coincide con lo guardado
    await page.reload();
    await expect(recuento).toHaveText('1 pendiente de 3 tareas');

    // 3. Reabre «Enviar la factura»: el texto vuelve a subir sin recargar la página
    await page.evaluate(() => { document.body.dataset.sinRecargar = 'true'; });
    await page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura' }).click();
    await expect(recuento).toHaveText('2 pendientes de 3 tareas');
    await expect(pagina).toHaveAttribute('data-sin-recargar', 'true');

    // 4. Al recargar, sigue igual
    await page.reload();
    await expect(recuento).toHaveText('2 pendientes de 3 tareas');
  });
});
