import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('reabrir una tarea', async ({ page, tasks }) => {
    await tasks([
      { title: 'Enviar la factura', completed: true },
      { title: 'Revisar el presupuesto', completed: true },
    ]);
    await page.reload();
    const factura = page.getByRole('listitem').filter({ hasText: 'Enviar la factura' });

    // 1. Pulsa el círculo de «Enviar la factura» y espera la respuesta del servidor
    const guardado = page.waitForResponse((res) => res.url().endsWith('/toggle'));
    await page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura' }).click();
    await guardado;

    // 2. Se ve pendiente al momento: sin tachar y sin «Completada»
    await expect(page.getByRole('button', { name: 'Completar tarea: Enviar la factura', pressed: false })).toBeVisible();
    await expect(factura.locator('.task-title')).toHaveCSS('text-decoration-line', 'none');
    await expect(factura.getByText('Completada', { exact: true })).toHaveCount(0);

    // 3. Al recargar sigue pendiente, y la otra tarea sigue completada
    await page.reload();
    await expect(page.getByRole('button', { name: 'Completar tarea: Enviar la factura', pressed: false })).toBeVisible();
    await expect(factura.locator('.task-title')).toHaveCSS('text-decoration-line', 'none');
    await expect(factura.getByText('Completada', { exact: true })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Reabrir tarea: Revisar el presupuesto', pressed: true })).toBeVisible();
  });
});
