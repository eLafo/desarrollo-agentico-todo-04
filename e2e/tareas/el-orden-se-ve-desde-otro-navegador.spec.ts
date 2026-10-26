import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('el orden se ve desde otro navegador', async ({ page, browser, tasks }) => {
    await tasks([{ title: 'Enviar la factura' }, { title: 'Revisar el presupuesto' }, { title: 'Llamar al proveedor' }]);
    await page.reload();

    // 1. Pulsa «Bajar tarea: Enviar la factura» y espera el guardado
    const guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await page.getByRole('button', { name: 'Bajar tarea: Enviar la factura' }).click();
    await guardado;

    // 2. Abre la lista en otro contexto, sin cookies ni almacenamiento compartidos
    const otro = await browser.newContext();
    const otraPagina = await otro.newPage();
    await otraPagina.goto(page.url());

    // 3. Allí la lista tiene el mismo orden
    await expect(otraPagina.locator('.task-title')).toHaveText(['Revisar el presupuesto', 'Enviar la factura', 'Llamar al proveedor']);
    await otro.close();
  });
});
