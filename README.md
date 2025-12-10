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
go install github.com/japablazatww/centralnexus/nexus/cmd/nexus-cli@develop

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

-   **Consumer**: Ejecutará pruebas contra el servidor.

## Flujo de Trabajo y Generación de Código

Es crucial entender cómo y cuándo se genera el código en este ecosistema distribuido.

### 1. ¿Cuándo se genera el código?
El código (**auto-generado**) reside en `nexus/generated`. Se regenera cuando:
*   Alguien actualiza una librería externa (e.g. `libreria-a`).
*   Ejecutas el comando:
    ```bash
    nexus-cli build
    ```
    *(Nota: La CLI detecta automáticamente la carpeta `nexus/generated` si estás en la raíz, o puedes usar `--output`.)*

### 2. ¿Cómo funciona "bajo el capó"?
1.  **Crawl & Index**: La CLI descarga las librerías (`go get`), las analiza (AST Parsing) y detecta funciones exportadas.
2.  **Catalogación**: Crea un `catalog.json` con metadatos (rutas, parámetros, tipos).
3.  **Generación de Templates**:
    *   **Server (`server_gen.go`)**: Crea "Adapters" que normalizan parámetros (Fuzzy Matching) y llaman a la librería real.
    *   **SDK (`sdk_gen.go`)**: Crea una jerarquía de clientes (`Client.Domain.Subdomain...`) para el consumidor.

### 3. Colaboración: ¿Cuándo pushear?
**SÍ, se debe pushear el código generado.**

Al trabajar en equipo:
1.  **Dev A** actualiza `libreria-a`.
2.  **Dev A** corre `nexus-cli build` localmente para regenerar los adaptadores.
3.  **Dev A** hace commit de los cambios en `nexus/generated`.
4.  **Dev B** hace `git pull` y ya tiene el servidor y SDK listos para usar sin necesidad de regenerar nada.

*Si hay conflicto en archivos generados, simplemente descarta tus cambios locales en `generated/` y corre `nexus-cli build` nuevamente.*

1.  Edita `nexus/cmd/nexus-cli`.
2.  Actualiza el registro en `nexus/cmd/nexus-cli/registry.json`.
3.  Reinstala localmente (`go install .`).
