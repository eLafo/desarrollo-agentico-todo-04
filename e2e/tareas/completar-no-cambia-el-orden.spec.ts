import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('completar no cambia el orden', async ({ page, tasks }) => {
    await tasks([
      { title: 'Enviar la factura' },
      { title: 'Revisar el presupuesto', completed: true },
      { title: 'Llamar al proveedor' },
    ]);
    await page.reload();
    const titulos = page.locator('.task-title');
    const orden = ['Revisar el presupuesto', 'Enviar la factura', 'Llamar al proveedor'];

    // 1. Pulsa «Subir tarea: Revisar el presupuesto»: la completada se mueve como las demás
    let guardado = page.waitForResponse((res) => res.url().endsWith('/move'));
    await page.getByRole('button', { name: 'Subir tarea: Revisar el presupuesto' }).click();
    await guardado;
    await expect(titulos).toHaveText(orden);

    // 2. Completa «Enviar la factura» y reabre «Revisar el presupuesto»: el orden no cambia
    guardado = page.waitForResponse((res) => res.url().endsWith('/toggle'));
    await page.getByRole('button', { name: 'Completar tarea: Enviar la factura' }).click();
    await guardado;
    guardado = page.waitForResponse((res) => res.url().endsWith('/toggle'));
    await page.getByRole('button', { name: 'Reabrir tarea: Revisar el presupuesto' }).click();
    await guardado;
    await expect(titulos).toHaveText(orden);

    // 3. Al recargar sigue el orden, con los estados cambiados
    await page.reload();
    await expect(titulos).toHaveText(orden);
    await expect(page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura', pressed: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Completar tarea: Revisar el presupuesto', pressed: false })).toBeVisible();
  });
});
