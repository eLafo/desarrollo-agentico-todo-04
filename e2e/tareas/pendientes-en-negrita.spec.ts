import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('el número de pendientes va en negrita', async ({ page, tasks }) => {
    await tasks([
      { title: 'Enviar la factura' },
      { title: 'Revisar el presupuesto' },
      { title: 'Pagar el alquiler', completed: true },
    ]);
    await page.reload();
    const recuento = page.locator('.list-count');
    const numero = recuento.locator('strong');
    const pagina = page.locator('body');

    // 1. Al cargar, solo el número de pendientes va en negrita
    await expect(recuento).toHaveText('2 pendientes de 3 tareas');
    await expect(numero).toHaveText('2');
    await expect(numero).toHaveCSS('font-weight', '700');
    await expect(recuento).toHaveCSS('font-weight', '400');

    // 2. Completa «Enviar la factura»: el número baja sin recargar y sigue en negrita
    await page.evaluate(() => { document.body.dataset.sinRecargar = 'true'; });
    await page.getByRole('button', { name: 'Completar tarea: Enviar la factura' }).click();
    await expect(recuento).toHaveText('1 pendiente de 3 tareas');
    await expect(numero).toHaveText('1');
    await expect(numero).toHaveCSS('font-weight', '700');
    await expect(recuento).toHaveCSS('font-weight', '400');
    await expect(pagina).toHaveAttribute('data-sin-recargar', 'true');

    // 3. Reabre «Enviar la factura»: el número vuelve a subir sin recargar y sigue en negrita
    await page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura' }).click();
    await expect(recuento).toHaveText('2 pendientes de 3 tareas');
    await expect(numero).toHaveText('2');
    await expect(numero).toHaveCSS('font-weight', '700');
    await expect(recuento).toHaveCSS('font-weight', '400');
    await expect(pagina).toHaveAttribute('data-sin-recargar', 'true');
  });
});
