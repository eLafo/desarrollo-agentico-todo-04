import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('completar una tarea', async ({ page, tasks }) => {
    await tasks([{ title: 'Enviar la factura' }, { title: 'Revisar el presupuesto' }]);
    await page.reload();
    const factura = page.getByRole('listitem').filter({ hasText: 'Enviar la factura' });

    // 1. Pulsa el círculo de «Enviar la factura» y espera la respuesta del servidor
    const guardado = page.waitForResponse((res) => res.url().endsWith('/toggle'));
    await page.getByRole('button', { name: 'Completar tarea: Enviar la factura' }).click();
    await guardado;

    // 2. Se ve completada al momento: tachada y con «Completada»
    await expect(page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura', pressed: true })).toBeVisible();
    await expect(factura.locator('.task-title')).toHaveCSS('text-decoration-line', 'line-through');
    await expect(factura.getByText('Completada', { exact: true })).toBeVisible();

    // 3. Al recargar sigue completada, y la otra tarea sigue pendiente
    await page.reload();
    await expect(page.getByRole('button', { name: 'Reabrir tarea: Enviar la factura', pressed: true })).toBeVisible();
    await expect(factura.locator('.task-title')).toHaveCSS('text-decoration-line', 'line-through');
    await expect(factura.getByText('Completada', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Completar tarea: Revisar el presupuesto', pressed: false })).toBeVisible();
  });
});
