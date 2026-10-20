---
name: write_plan
description: Escribe el plan de un cambio de código, o una revisión del plan vigente, a partir de una especificación clara.
skills: [feature, bug, chore]
---

# write_plan

## Entradas

- `specification`: título, descripción, aclaraciones y tipo de la petición.
- `change_request`: peticiones de cambio, si las hay: las de una revisión del humano, con los comentarios que las motivan, o el rechazo de la validación, con su acta y los defectos de cada dimensión.
- Los planes anteriores, si los hay.
- La ruta donde escribir el plan, que da el adaptador.

## Qué hace

1. Prepara el encargo para la skill del tipo (`.claude/skills/<tipo>/SKILL.md`): la especificación y, si hay peticiones de cambio, esas peticiones con lo que las motiva, el plan vigente como contexto y el aviso de que la rama, o la copia de trabajo si la validación lo ha rechazado, ya lo implementa. Un rechazo de la validación se corrige planificando el arreglo de cada defecto; no pidas cambiar las afirmaciones que lo detectaron. La skill no sabe de revisiones: recibe un encargo y escribe un plan completo en la ruta que da el adaptador.
2. Ejecuta la skill. Planifica leyendo el código: no ejecutes, depures ni modifiques la aplicación. Si encuentras un problema ajeno al encargo, anótalo como observación y no lo investigues.
3. Si encuentras una duda funcional que la especificación no resuelve, no preguntes: devuélvela como resultado.
4. Comprueba que el plan existe y que cubre la especificación y las peticiones, y que su sección «Documentación afectada» nombra los documentos que el cambio deja desfasados o dice por qué no hay ninguno.
5. Si es una revisión, encabeza el plan con qué cambia respecto a la revisión anterior y qué peticiones recoge, citando su origen: «revisión del humano» o «rechazo de validación, intento k», con los defectos de cada dimensión.
6. Publica el plan para informar, no para que lo aprueben. A la vista, lo que recoge, qué cambia respecto a la revisión anterior, el enfoque, los ficheros que toca y los títulos de los pasos; plegado, el plan completo.

## Resultado

El plan o la revisión publicados, o una duda funcional que la especificación no resuelve.
