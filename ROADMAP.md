# Roadmap

Mejoras del sistema `adw` que quedan para más adelante. Cada una dice qué problema resuelve, en qué consiste y qué hay que decidir antes de hacerla.

## El verificador escribe los tests de aceptación

**Problema.** Muchas afirmaciones del plan de verificación se comprueban con pruebas sueltas que el verificador ejecuta a mano (peticiones HTTP, consultas a la base, medidas del DOM), y desaparecen al cerrar el caso. En la [issue #22 de `desarrollo-agentico-todo-04`](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/22), el acta lo dice: «Ningún test automático protege que las tareas de 121 a 200 caracteres se conserven». Los tests que se quedan en el repositorio los escribe quien implementa. El verificador solo comprueba que existen, que fallan sin el cambio y que cubren el criterio.

**En qué consiste.** El verificador trabaja sin el plan de implementación. Convierte en tests de aceptación lo que se puede comprobar con código y los registra en la rama del caso antes de implementar. Como el cambio aún no existe, fallan sobre la base, y la comprobación en rojo y en verde sale sola. La implementación los pone en verde sin modificarlos, y la validación comprueba que siguen intactos. La validación sigue en local: la CI es una comprobación más en el servidor, que da visibilidad en el PR.

**Qué hay que decidir:**

- **El commit en la rama antes de implementar.** Hoy `plan_verification` no toca la copia de trabajo: lee la base con `git show origin/<base>:<ruta>` y corre en paralelo con `setup_environment` y `write_plan`. Con este cambio necesita el entorno listo y un commit propio, en la rama del caso. Hay que fijar el mensaje del commit y ver si se mantiene el paralelo con `write_plan`.
- **Cómo se comprueba que la implementación no ha tocado los tests.** Por ejemplo, con una afirmación bloqueante de fidelidad: `git diff <commit del verificador> HEAD -- <tests de aceptación>` vacío.
- **Qué pasa si un test de aceptación está mal.** Solo lo cambia una revisión del plan de verificación, nunca `implement`. Hay que ver cómo llega ese defecto al verificador dentro del bucle de corrección.
- **Qué se queda como prueba suelta.** Lo que no merece un test permanente, y la demostración para el revisor (el vídeo y las capturas), que enseña el resultado pero no lo valida.
