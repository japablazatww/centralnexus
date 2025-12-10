# Arquitectura y Lógica Interna de Nexus

Este documento técnico explica en detalle cómo funciona CentralNexus "bajo el capó". Está diseñado para desarrolladores que desean entender la magia detrás del autodescubrimiento y la generación de código.

## 1. El Concepto Fundamental

Nexus no es solo un servidor HTTP; es un **Compilador de Integración**.

Su objetivo es tomar código fuente Go existente (librerías desconectadas) y convertirlas automáticamente en:
1.  Una API HTTP unificada.
2.  Un SDK cliente tipado y estructurado.

Todo esto sin escribir archivos `.proto` ni configuraciones manuales complejas. La fuente de verdad es **el código de la librería**.

---

## 2. El Cerebro: `nexus-cli`

La pieza central es la herramienta de línea de comandos `nexus-cli`. Su comando principal, `nexus-cli build`, orquesta todo el ciclo de vida.

### Ciclo de Vida de `build`

Cuando ejecutas `nexus-cli build`, ocurren 4 fases secuenciales:

#### Fase A: Ingestión y Aislamiento
1.  **Lectura del Registro**: Lee `registry.json` para saber qué librerías procesar (e.g., `github.com/japablazatww/libreria-a`).
2.  **Entorno Temporal**: Crea un directorio temporal (e.g., `/tmp/nexus-build-xyz`) y ejecuta `go mod init`.
3.  **Descarga**: Ejecuta `go get package@branch` (actualmente forzado a `@develop`). Esto descarga el código fuente real de las librerías al entorno temporal.

#### Fase B: Análisis de Código (AST Parsing)
Aquí ocurre la magia. Nexus no importa la librería para ejecutarla todavía; la **lee**.
1.  **Exploración Recursiva**: Comienza en la raíz de la librería y busca archivos `lib_config.json`.
    *   Si encuentra uno, sabe que es un **Dominio** (e.g., "System").
    *   Si este config tiene `nested_domains`, desciende a las subcarpetas (e.g., "Transfers" -> "National").
2.  **Parsing de Go**: Usa el paquete `go/ast` (Abstract Syntax Tree) para leer los archivos `.go`.
    *   Identifica funciones exportadas (que empiezan con Mayúscula).
    *   Extrae nombres de parámetros y tipos de retorno.
    *   Extrae comentarios de documentación.
3.  **Result**: Construye una estructura de datos en memoria (`Catalog`) que representa todo el árbol de funciones de la librería.

#### Fase C: Generación de Código (Templates)
Con el `Catalog` en memoria, Nexus escribe dos archivos Go críticos en tu carpeta `nexus/generated`.

**1. `server_gen.go` (El Servidor)**
*   **Routing**: Crea un `mux.HandleFunc` por cada función descubierta (e.g., `libreria-a.system.GetSystemStatus`).
*   **Adapters**: Crea funciones "wrapper" que:
    *   Reciben un JSON genérico.
    *   **Normalizan Parámetros**: Aplican lógica "Fuzzy" (e.g., si llega `user_id`, busca `UserID` o `UserId` en los inputs reales).
    *   Invocan la función real de la librería importada.
    *   Devuelven la respuesta en JSON.

**2. `sdk_gen.go` (El Cliente)**
*   **Árbol de Tipos**: Crea structs anidados que imitan la estructura de carpetas de la librería.
    ```go
    type Client struct {
        Libreriaa *LibreriaaClient
    }
    type LibreriaaClient struct {
        System *LibreriaaSystemClient // Apunta a los métodos de system/
        ...
    }
    ```
*   **Métodos**: Cada struct tiene métodos (e.g., `GetSystemStatus`) que hacen el `POST` HTTP al servidor con la ruta correcta.

#### Fase D: Compilación Final
El desarrollador (tú) compila el proyecto (`go build`). Como `server_gen.go` y `sdk_gen.go` son archivos Go válidos que importan las librerías reales, el compilador de Go verifica que todo coincida (tipos, nombres, etc.).

---

## 3. Flujo de Ejecución (Runtime)

Una vez compilado y corriendo (`nexus.exe` y `consumer.exe`), el flujo es:

1.  **Consumer**: Llama a `client.Libreriaa.Transfers.National.Transfer(...)`.
2.  **SDK**:
    *   Serializa los argumentos a JSON `{"source_account": "...", ...}`.
    *   Hace POST a `http://host:8080/libreria-a.transfers.national.Transfer`.
3.  **Nexus Server**:
    *   Recibe el Request.
    *   Enruta al handler generado en `server_gen.go`.
    *   **Adapter**:
        *   Desempaqueta el JSON.
        *   Busca los parámetros (usando fuzzy match si es necesario).
        *   Llama a `libreria_a_transfers_national.Transfer(...)` (código original de la librería).
    *   Serializa la respuesta de la función.
4.  **Consumer**: Recibe el JSON y lo deserializa en la estructura de respuesta (o `interface{}`).

## 4. Preguntas Frecuentes

### ¿Qué pasa si instalo una nueva dependencia?
1.  Añádela a `registry.json` en `nexus-cli`.
2.  Corre `nexus-cli build`.
3.  Nexus la descargará, la escaneará y regenerará los archivos `_gen.go` para incluirla automáticamente.

### ¿Por qué `nexus-cli build` tarda un poco?
Porque está haciendo un `go get` real y analizando código fuente. Es un proceso de compilación/transpilación.

### ¿Por qué pushear la carpeta `nexus/generated`?
Porque esos archivos son el puente. Sin ellos, el servidor no sabe qué endpoints exponer y el cliente no sabe qué métodos existen. Al versionarlos, garantizas que cualquier clon del repo funcione inmediatamente.
