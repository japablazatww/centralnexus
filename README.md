# Prueba de Concepto CentralNexus

Esta PoC demuestra la arquitectura CentralNexus, unificando librerías bajo un servicio centralizado que expone un **SDK Genérico Dinámico** y una CLI "Smart".

## Características Clave

1.  **Smart CLI (`nexus-cli`)**: Herramienta unificada que descubre, instala (`go get`) e indexa automáticamente las librerías soportadas sin configuración.
2.  **SDK con Namespacing**: El cliente accede a las librerías de forma organizada (e.g., `client.LibreriaA.Method()`).
3.  **Proxy Dinámico**: El servidor Nexus mapea automáticamente los nombres de parámetros (Camel/Snake/Pascal case).

## Estructura

- **centralnexus**: Monorepo principal.
  - **nexus**: Servidor Central y CLI.
    - `cmd/nexus-cli`: Herramienta todo-en-uno (builder + search).
  - **consumer**: Cliente de ejemplo.

## Guía Rápida

### 1. Nexus CLI (Descubrimiento)

La CLI se autogestiona. No necesitas generar catálogos manualmente.

```bash
# Instalación
$env:GOPROXY="direct"
go install github.com/japablazatww/centralnexus/nexus/cmd/nexus-cli@latest

# Uso (Auto-descubrimiento en la primera ejecución)
nexus-cli search --search-param user_id
```

### 2. Ejecutar Servidor y Consumidor (Docker)

Para ver la integración completa funcionando:

```bash
docker-compose up --build
```

Esto levantará:
-   **Nexus Server**: Escuchando en puerto 8080.
-   **Consumer**: Ejecutará pruebas contra el servidor.

## Desarrollo

Si deseas modificar la lógica de generación:

1.  Edita `nexus/cmd/nexus-cli`.
2.  Actualiza el registro en `nexus/cmd/nexus-cli/registry.json`.
3.  Reinstala localmente (`go install .`).
