# Guía de Pruebas con Docker Compose

Esta guía explica cómo levantar todo el entorno CentralNexus y ejecutar las pruebas de integración automáticamente.

## ¿Qué se prueba?
El entorno levanta dos contenedores:
1.  **nexus-server**: El sistema central que carga la `libreria-a` y expone la API HTTP.
2.  **consumer**: Un cliente Go que utiliza el `SDK Genérico` para llamar a 3 servicios simulados:
    *   `GetSystemStatus`: Verifica conectividad.
    *   `GetUserBalance`: Prueba el mapeo de parámetros (Snake vs Camel Case).
    *   `Transfer`: Prueba una operación transaccional.

## Instrucciones Rapidas

### 1. Prerrequisito (Opcional)
**Solo si has modificado código en `libreria-a` o agregado nuevas librerías**, necesitas regenerar el código antes de levantar Docker.

```bash
nexus-cli build
```

> **Nota Importante sobre Rutas**:
> Si ejecutas esto desde la raíz del proyecto, la CLI detectará automáticamente la carpeta `nexus/generated`.
> Si prefieres ejecutarlo desde otro lugar (o si la autodetección falla), **debes especificar dónde guardar el código** usando el flag `--output`:
> ```bash
> nexus-cli build --output D:\banca\centralnexus_poc\centralnexus\nexus\generated
> ```
> El objetivo es siempre actualizar los archivos que Docker va a consumir.
```

Si solo estás probando el repositorio tal cual lo descargaste, salta al paso 2. El código generado ya está versionado.

### 2. Levantar y Testear
Ejecuta el siguiente comando en la raíz del proyecto (`centralnexus_poc`):

```bash
docker-compose up --build --abort-on-container-exit
```

*Nota: `--abort-on-container-exit` hará que Docker se detenga automáticamente cuando el consumidor termine sus pruebas.*

### 2. Verificar Resultados

Busca en la consola la salida del contenedor `consumer`. Deberías ver algo como esto:

```text
consumer-1  | --- Testing GetSystemStatus ---
consumer-1  | System Status: OPERATIONAL
consumer-1  | 
consumer-1  | --- Testing GetUserBalance (CamelCase/SnakeCase check) ---
consumer-1  | Balance: 1000.5
consumer-1  | 
consumer-1  | --- Testing Transfer ---
consumer-1  | Transfer ID: TX-123456789
consumer-1  | Is Exited
```

Si ves estos 3 resultados, **el sistema funciona correctamente**:
1.  La CLI/Builder generó bien el código.
2.  El Server está ruteando las peticiones.
3.  El SDK está enviando los parámetros correctamente.

### 3. Limpieza
Para borrar los contenedores creados:
```bash
docker-compose down
```
