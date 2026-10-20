import { test, expect } from '../fixtures';

test.describe('Tareas', () => {
  test('el botón de añadir muestra «Añadir tarea» al pasar el ratón', async ({ page }) => {
    const formulario = page.getByRole('form', { name: 'Añadir tarea', exact: true });

    // 1. El botón del formulario de nueva tarea lleva el title «Añadir tarea», el texto que el
    //    navegador muestra al pasar el ratón
    const boton = formulario.getByRole('button');
    await expect(boton).toHaveAttribute('title', 'Añadir tarea');
  });
});
