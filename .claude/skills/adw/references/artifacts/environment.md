# Environment

El lugar donde se trabaja en el caso: una rama propia, en una copia de trabajo propia, y un entorno de desarrollo aislado, con su base de datos y sus puertos, que se arranca cuando hace falta y se para al terminar.

## Información

- La rama del caso. Se asigna una sola vez, a partir de la petición de trabajo, y no cambia.
- La rama base de la que parte.
- Si la copia de trabajo está en la rama del caso y si el entorno de desarrollo está en marcha.

## Contrato

- **Entradas:**
  - `work_item`: la petición de trabajo (identidad, tipo y título), y si está clara.
  - `demand`: si queda trabajo que necesita el entorno en marcha.
- **Expone:** `ready`, `released` y la rama del caso.

## Ciclo de vida

| Stage | Guard | Servicio | Milestones |
| --- | --- | --- | --- |
| `Setup` | El `work_item` está claro, hay `demand` y el Environment no tiene `ready` | [`setup_environment`](../services/setup_environment.md) | `ready` |
| `Teardown` | Con `ready` y sin `demand` | [`teardown_environment`](../services/teardown_environment.md) | `released` |

| Milestone | Se alcanza cuando | Se invalida cuando |
| --- | --- | --- |
| `ready` | La copia de trabajo está en la rama del caso, al día, y el entorno de desarrollo responde | La copia de trabajo cambia de rama, se queda atrás o el entorno deja de responder |
| `released` | El entorno de desarrollo está parado y la copia de trabajo sigue en la rama del caso | El entorno vuelve a arrancar o la copia de trabajo cambia de rama |

`ready` y `released` describen la copia de trabajo del caso en la máquina donde trabaja el sistema, no el caso en general. Cada caso tiene la suya y su entorno, que no comparte con ningún otro: trabajar en otro caso no los invalida, y varios casos pueden avanzar a la vez. La copia de trabajo existe antes de la primera invocación del caso y se borra cuando el caso termina; quién lo hace lo dice el adaptador.
