---
description: Stop the TODO development environment without removing containers, volumes or data
allowed-tools: Bash, Read
---

# Para el entorno de desarrollo

Detén solo lo que pertenezca a esta copia de trabajo: el checkout principal o un worktree de `bin/tree` (`trees/issue-N`). No elimines contenedores, volúmenes ni datos, y no uses `docker compose down`.

## Puertos

Los de la sesión, con `printenv APP_PORT DB_PORT`. Si no están y la copia de trabajo no tiene `.env` (`git status --short --ignored .env` no muestra `!! .env`), es el checkout principal: `3000` y `5432`. Si tiene `.env` y no están, el entorno no está cargado: detente y dilo. Escribe los puertos tal cual en los comandos.

## Detecta el modo

- **Docker completo:** `docker compose ps` muestra el servicio `app` en marcha.
- **Host:** hay un proceso `buffalo dev` cuyo directorio de trabajo es la raíz de esta copia de trabajo.

Si no hay nada en marcha, informa de ello y termina.

## Modo Docker completo

1. Ejecuta `docker compose stop`.
2. Comprueba con `docker compose ps` que ya no queda ningún servicio en marcha.

## Modo host

1. Localiza los procesos `buffalo dev` con `pgrep -f "buffalo dev"`. Para cada uno, comprueba con `lsof -a -p <pid> -d cwd` que su directorio de trabajo es la raíz de esta copia de trabajo. Si pertenece a otro proyecto o a otro worktree, no lo toques.
2. Detén los de esta copia de trabajo con `kill <pid>`.
3. Ejecuta `docker compose stop db`.

## Comprobación

`http://localhost:<puerto de la aplicación>/healthz` ya no responde y nada de esta copia de trabajo sigue escuchando en sus puertos.

## Resultado

Informa de:

- modo detectado;
- servicios y procesos detenidos;
- resultado de la comprobación;
- procesos de otros proyectos o worktrees que hayas dejado intactos.
