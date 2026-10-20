---
name: document
description: Escribe o actualiza el documento de cierre de un cambio en docs/history/changes/ y su entrada en el índice condicional, a partir del diff real y del acta de su verificación, y registra en docs/history/decisions/ las preguntas con opciones que respondió el humano. Úsala para documentar el cambio de una rama antes de publicarlo. No registra ni publica nada.
allowed-tools: Read, Glob, Grep, Write, Edit, Bash(git diff *), Bash(git status *), Bash(git log *), Bash(git show *)
---

# Documenta un cambio

El documento de cierre explica qué se construyó, cómo se usa, cómo se probó y qué conviene saber para trabajar sobre ello, para quien llegue dentro de seis meses. Se escribe en la rama y viaja en el PR con el código.

Es la historia del caso, y no sustituye a la documentación viva: lo que hace el sistema y cómo se trabaja en él están en `docs/knowledge/` y `docs/procedures/`, y los actualiza el propio cambio. No los toques: si ves uno desfasado, dilo en tu informe.

## Entradas

- La petición: número, tipo, título y criterios.
- La base del cambio, por ejemplo `origin/main`.
- El acta de la verificación: su enlace y su resumen.
- La ruta del documento, `docs/history/changes/<tipo>-<N>-<slug>.md`.
- Las preguntas con opciones que respondió el humano en el caso, con las opciones, su respuesta y el enlace a ella. Pueden no ser ninguna.

## Procedimiento

1. Lee el cambio real: `git diff --stat <base>`, `git diff <base>` en los ficheros que importan y los ficheros nuevos que lista `git status --porcelain`. Documentas lo que hay, no lo que decía el plan ni lo que recuerdas.
2. Si el documento ya existe, actualízalo para que describa el diff vigente. No crees otro.
3. Escribe el documento con esta estructura. Resumen, Implementación técnica y Cómo se probó son obligatorias; el resto se adapta al tamaño del cambio:

   ```markdown
   # <Título del cambio>

   **Issue:** #<N> · **Tipo:** <tipo>

   ## Resumen
   <Dos o tres frases: qué se construyó y por qué.>

   ## Qué se construyó
   <Las capacidades que añade o cambia, según el diff.>

   ## Implementación técnica
   <Los ficheros clave con su papel y los tres a cinco cambios técnicos que importan.>

   ## Cómo se usa
   <El recorrido del usuario o la forma de usarlo.>

   ## Configuración
   <Variables, ajustes o migraciones. Si no hay: «No requiere configuración.».>

   ## Cómo se probó
   <Las afirmaciones verificadas y su resultado, con el enlace al acta, y los comandos para repetir las pruebas.>

   ## Notas
   <Decisiones, limitaciones y lo que queda fuera.>
   ```

4. Añade su entrada al índice `docs/history/changes/README.md`, o actualízala si ya existe, con tres a cinco condiciones de carga concretas, en el formato que describe el propio índice.
5. Si el humano respondió preguntas con opciones, escribe `docs/history/decisions/<tipo>-<N>-<slug>.md`, con el mismo nombre que el documento de cierre, y su entrada en `docs/history/decisions/README.md`, en el formato que describe ese índice. Una sección por pregunta, con la respuesta en palabras del humano. Enlázalo desde las Notas del documento de cierre. Si no hubo preguntas, no lo crees.
6. Comprueba que solo has escrito esos ficheros: el documento de cierre y su entrada y, si las hay, las decisiones y su entrada.

## Reglas

- Concisión: decisiones y uso, no una paráfrasis del diff. Prefiere listas y tablas a párrafos largos.
- Las condiciones del índice nombran rutas, modelos, pantallas o problemas concretos del cambio, para que quien planifique otro trabajo sepa si le toca leerlo.
- No registres ni publiques los cambios: lo hace quien te invoca.
