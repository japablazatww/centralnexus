# Guía de Usuario: Nexus Smart CLI

Esta herramienta permite descubrir servicios integrados en CentralNexus de forma automática, sin configuración manual.

## 1. Instalación

Como el código está alojado en GitHub, puedes instalar la herramienta directamente con Go.

### Windows (PowerShell)
```powershell
$env:GOPROXY="direct"
go install github.com/japablazatww/centralnexus/nexus/cmd/nexus-cli@develop
```

### Linux / macOS
```bash
GOPROXY=direct go install github.com/japablazatww/centralnexus/nexus/cmd/nexus-cli@develop
```

*Nota: `GOPROXY=direct` asegura que descargues la última versión, ignorando el caché.*

## 2. Uso Básico (Zero Config)

La CLI es "inteligente". Al ejecutarla por primera vez, detectará que no tiene datos y automáticamente:
1.  Descargará las librerías soportadas.
2.  Analizará el código fuente.
3.  Generará un catálogo local en tu carpeta de usuario (`~/.nexus/catalog.json`).

### Buscar un Parámetro
```bash
nexus-cli search --search-param user_id
```

**Salida Ejemplo:**
```text
Found 1 services with parameter 'user_id':
- libreria-a.GetUserBalance
  Match: user_id (Input)
```

## 3. Comandos Disponibles

### `search`
Busca servicios que contengan un parámetro específico (flexible: acepta snake_case, camelCase, PascalCase).

```bash
nexus-cli search --search-param [nombre]
```

### `build`
Fuerza la regeneración del catálogo. Útil si sabes que las librerías se han actualizado y quieres refrescar tu índice local.

```bash
nexus-cli build
```

### `dump-catalog`
Imprime el JSON crudo del catálogo actual. Útil para verificar qué datos tiene la herramienta.

```bash
nexus-cli dump-catalog
```

## 4. Solución de Problemas (Debugging)

Si la herramienta no encuentra lo que esperas o falla, usa la bandera `--debug` para ver qué está pasando "bajo el capó".

```bash
nexus-cli search --search-param user --debug
```

Esto mostrará:
-   Rutas exactas donde busca el catálogo.
-   Logs de descarga de librerías (`go get`).
-   Detalles del análisis de código (archivos visitados, funciones encontradas).

## 5. Generación de Código (Under the Hood)

Cuando ejecutas `nexus-cli build`, la herramienta no solo indexa, sino que **escribe código Go** en la carpeta `nexus/generated`.

### Archivos Generados
1.  **`server_gen.go` (Server Adapters)**:
    -   Contiene los `http.Handler` para cada función descubierta.
    -   Implementa la lógica de **Fuzzy Matching** para parámetros (e.g. `user_id` -> `userID`).
    -   Actúa como puente entre la API HTTP y la librería real.

2.  **`sdk_gen.go` (Client SDK)**:
    -   Genera una estructura de cliente jerárquica basada en los dominios DDD.
    -   Permite al consumidor usar autocompletado: `client.LibreriaA.Transfers.National...`.
    -   Abstrae las llamadas HTTP y la serialización JSON.

### ¿Dónde debo ejecutar `nexus-cli build`?
La herramienta tiene detección inteligente de la carpeta `nexus/generated`.

**Opciones:**
1.  **Automático**: Ejecuta desde cualquier subcarpeta conocida (`root`, `nexus/`, `nexus/cmd/`...). La herramienta buscará `generated` hacia arriba o abajo.
2.  **Manual**: Si no la encuentra, usa el flag `--output`:

```bash
nexus-cli build --output /ruta/absoluta/a/nexus/generated
```

Esto te da flexibilidad total para ejecutar la CLI desde donde quieras.

*Recuerda: Si algo cambia en las librerías base, corre `build` para actualizar estos archivos.*
