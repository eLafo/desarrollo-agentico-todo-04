# Visión general

Una aplicación web con una sola lista de tareas, sin usuarios ni cuentas. Se pueden añadir tareas, editar su título, marcarlas como completadas o reabrirlas y eliminarlas. Lo que hace cada pantalla está en [features/tareas.md](features/tareas.md).

## Stack

| Pieza | Qué es |
| --- | --- |
| Lenguaje | Go 1.26 (`go.mod`) |
| Framework | Buffalo 1.1, con Pop 6 como ORM y plantillas Plush |
| Base de datos | PostgreSQL 17.6, en Docker (`compose.yaml`) |
| Interfaz | HTML renderizado en el servidor, con CSS y JavaScript propios y sin dependencias (`public/assets/`) |
| Tests | Suites de Buffalo para Go y Playwright para el navegador |

## Estructura

| Ruta | Contenido |
| --- | --- |
| `cmd/app/` | El arranque: aplica las migraciones pendientes y sirve la aplicación |
| `actions/` | Rutas y middleware (`app.go`), acciones de las tareas (`tasks.go`) y comprobación de salud (`health.go`) |
| `models/` | La conexión a la base de datos, las migraciones (`models.go`) y el modelo `Task` con su validación (`task.go`) |
| `templates/` | La plantilla común (`application.plush.html`), los avisos (`_flash.plush.html`) y las pantallas de tareas (`tasks/`) |
| `public/assets/` | `app.css` y `app.js` |
| `migrations/` | El esquema y los datos de ejemplo, embebidos en el binario |
| `fixtures/` | Datos para los tests de Go |
| `e2e/` | Tests de navegador |
| `adws/`, `config/adw.yml` | El lanzador del workflow `adw`. No forma parte de la aplicación |

Las capas y sus reglas están en [../procedures/conventions.md](../procedures/conventions.md).

## Datos

Una tabla, `tasks`:

| Columna | Tipo | Notas |
| --- | --- | --- |
| `id` | uuid | Clave primaria |
| `title` | string(200) | Obligatorio; hasta 200 caracteres (`models.TitleMaxLength`) |
| `completed` | bool | `false` por defecto |
| `created_at`, `updated_at` | timestamp | Los pone Pop |

Una migración de datos crea tres tareas de ejemplo, una de ellas completada. Solo las crea si no existen.

## Rutas

| Método | Ruta | Acción | Qué hace |
| --- | --- | --- | --- |
| `GET` | `/` | `TasksIndex` | La lista y el formulario para añadir |
| `POST` | `/tasks` | `TasksCreate` | Crea una tarea |
| `GET` | `/tasks/{task_id}/edit` | `TasksEdit` | El formulario de edición |
| `PUT` | `/tasks/{task_id}` | `TasksUpdate` | Cambia el título |
| `PUT` | `/tasks/{task_id}/toggle` | `TasksToggle` | Alterna entre completada y pendiente |
| `DELETE` | `/tasks/{task_id}` | `TasksDestroy` | Elimina la tarea |
| `GET` | `/healthz` | `HealthCheck` | `200 {"status":"ok"}` si responde la base de datos; si no, `503` |

Los formularios envían `PUT` y `DELETE` como `POST` con el campo `_method`.

## Recorrido de una petición

1. La ruta de `actions/app.go` elige la acción. Antes pasan los middleware: redirección a HTTPS (solo en producción), registro de parámetros, protección CSRF (Buffalo la desactiva con `GO_ENV=test`) y una transacción de base de datos por petición.
2. La acción toma la transacción (`c.Value("tx")`), lee la entrada y llama al modelo. `Task.Validate` decide si la entrada es válida.
3. Si hay errores de validación, la acción vuelve a mostrar el formulario con `422`. Si todo va bien, añade un aviso (flash) y redirige con `303` a `/`.
4. La pantalla se renderiza dentro de `application.plush.html`, que muestra los avisos pendientes y carga `app.css` y `app.js`.
5. `app.js` mejora lo que ya funciona sin él: pide confirmación antes de eliminar y cambia el estado de una tarea sin recargar la página.

## Configuración

| Variable | Para qué | Por defecto |
| --- | --- | --- |
| `GO_ENV` | El entorno: `development`, `test` o `production`. Elige la base de datos de `database.yml` | `development` |
| `DB_HOST` | El servidor de PostgreSQL en desarrollo y test | `127.0.0.1` |
| `DB_PORT` | Su puerto | `5432` |
| `DATABASE_URL` | La conexión en producción | Una local a `todo_production` |
| `ADDR`, `PORT` | Dónde escucha la aplicación | Los de Buffalo; en la imagen de Docker, `0.0.0.0` y `3000` |

Cómo se arranca y se prueba, también varios entornos a la vez, está en [../procedures/development.md](../procedures/development.md).
