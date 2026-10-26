import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('mover una tarea con el teclado', async ({ page, tasks }) => {
    await tasks([{ title: 'Enviar la factura' }, { title: 'Revisar el presupuesto' }, { title: 'Llamar al proveedor' }]);
    await page.reload();
    const titulos = page.locator('.task-title');
    const bajar = page.getByRole('button', { name: 'Bajar tarea: Enviar la factura' });
    const subir = page.getByRole('button', { name: 'Subir tarea: Enviar la factura' });

    // 1. Pone el foco en «Bajar tarea: Enviar la factura» y pulsa Intro
    await bajar.focus();
    let guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await page.keyboard.press('Enter');
    await guardado;
    await expect(titulos).toHaveText(['Revisar el presupuesto', 'Enviar la factura', 'Llamar al proveedor']);
    await expect(bajar).toBeFocused();

    // 2. Pulsa Intro otra vez: la tarea baja otro puesto
    guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await page.keyboard.press('Enter');
    await guardado;
    await expect(titulos).toHaveText(['Revisar el presupuesto', 'Llamar al proveedor', 'Enviar la factura']);

    // 3. Vuelve a «Subir tarea: Enviar la factura» con Mayús+Tabulador y pulsa la barra espaciadora
    await page.keyboard.press('Shift+Tab');
    await expect(subir).toBeFocused();
    guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await page.keyboard.press('Space');
    await guardado;
    await expect(titulos).toHaveText(['Revisar el presupuesto', 'Enviar la factura', 'Llamar al proveedor']);

    // 4. Al recargar sigue el mismo orden
    await page.reload();
    await expect(titulos).toHaveText(['Revisar el presupuesto', 'Enviar la factura', 'Llamar al proveedor']);
  });
});
