import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('el botón de añadir dice «Añadir tarea»', async ({ page }) => {
    const formulario = page.getByRole('form', { name: 'Añadir tarea', exact: true });

    // 1. El botón del formulario de nueva tarea dice exactamente «Añadir tarea»
    const boton = formulario.getByRole('button');
    await expect(boton).toHaveText('Añadir tarea');
    await expect(boton).toHaveAccessibleName('Añadir tarea');

    // 2. El campo sigue igual: «Nueva tarea», con el texto de ayuda «Añadir una tarea…»
    const campo = formulario.getByRole('textbox', { name: 'Nueva tarea', exact: true });
    await expect(campo).toBeVisible();
    await expect(campo).toHaveAttribute('placeholder', 'Añadir una tarea…');
  });
});
