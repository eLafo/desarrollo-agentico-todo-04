# Control

Además del estado GSM de sus artefactos, cada caso tiene tres datos de control. No forman parte del dominio: dicen si el sistema puede actuar sobre el caso. Cómo se representan en cada servicio lo dice el adaptador de la Issue.

## Admisión

Un caso entra en el sistema cuando su Issue está admitida. Si no lo está, el motor ignora cualquier evento del caso. Si el humano retira la admisión, el sistema suelta el caso.

## Propiedad

Un caso admitido pertenece a un único sistema. Un sistema solo actúa sobre casos libres o propios. Si el caso está libre, el motor lo reclama antes de abrir ningún stage; si al reclamarlo descubre otro dueño, renuncia y no actúa. La propiedad dura todo el caso, y la libera el humano.

## Turno

Cuando el caso queda en quiescencia, o cuando un stage falla o agota su tope, el motor cede el turno al humano y le dice qué espera de él y cómo devolverlo. Antes de cederlo deja la copia de trabajo limpia: si quedan cambios sin registrar, como una implementación que la verificación ha rechazado, los registra en la rama del caso. Y si la rama tiene trabajo sin validar, registrado ahora o antes (por ejemplo, al recuperar una ejecución que murió), lo publica como trabajo sin validar, para que el humano pueda verlo. Eso no es publicar la propuesta para revisión. Un caso parado nunca se queda con el turno del motor: si el motor no tiene nada que hacer, porque no encuentra nada nuevo o porque el caso ya ha terminado, cede el turno igualmente. Mientras el humano no lo devuelva, el motor no actúa sobre el caso aunque lleguen eventos. Al devolverlo, el motor reevalúa el caso con todo lo que el humano haya escrito. Tras un fallo, devolver el turno es reintentar, y los topes vuelven a cero.

## Peticiones directas

Un humano también puede dirigirse al sistema directamente, sin pasar por el servicio: «quiero hacer la issue #12», «he comentado en la issue X». Una petición directa sobre un caso lo admite si no lo estaba y, si el turno era del humano, se lo devuelve.

## Cuándo actúa el motor

Solo si el caso está admitido, es libre o propio, el turno no es del humano y no hay ningún stage abierto.
