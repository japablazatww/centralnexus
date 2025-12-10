# Estándar de Diseño de Librerías Nexus

Para que una librería sea compatible con el ecosistema Nexus y su CLI "Smart", debe cumplir con ciertos lineamientos estructurales y de código. Este documento define el **Contrato de Integración**.

## 1. Filosofía: Domain-Driven Design (DDD)

Nexus espera que las librerías no sean una "bolsa de funciones" plana. Espera una jerarquía semántica.

### Estructura de Directorios Esperada
```text
libreria-a/
├── lib_config.json          # (Opcional) Config raíz
├── system/                  # Dominio "System"
│   ├── lib_config.json      # Config del dominio
│   ├── functions.go         # Lógica exportada
│   └── structures.go        # Tipos de datos
└── transfers/               # Dominio "Transfers"
    ├── lib_config.json      # Config con nested_domains
    ├── national/            # Sub-dominio "National"
    │   ├── lib_config.json
    │   └── functions.go
    └── international/       # Sub-dominio "International"
        ├── lib_config.json
        └── functions.go
```

## 2. El Archivo `lib_config.json`

Es la "cédula de identidad" de cada carpeta. Nexus escanea recursivamente buscando este archivo.

**Formato:**
```json
{
  "is_domain": true,           // "true" si esta carpeta expone funciones directamente
  "has_nested_domains": true,  // "true" si tiene subcarpetas que son dominios
  "domains": ["national", "international"] // Lista de subcarpetas a escanear
}
```

*   Si una librería existente no tiene estos archivos, Nexus **NO** la verá.
*   **Refactor Required**: Las librerías legacy deben agregar estos archivos para ser descubiertas.

## 3. Convenciones de Código Go

Nexus analiza el AST (Abstract Syntax Tree) de Go. Para que una función sea expuesta como API, debe seguir estas reglas:

### A. Funciones Exportadas
Solo las funciones que inician con **Mayúscula** son visibles.
*   `func Transfer(...)` -> **VISIBLE** (API Endpoint generado).
*   `func calculateTax(...)` -> **INVISIBLE** (Lógica interna).

### B. Firmas de Funciones (Inputs/Outputs)
Nexus soporta tipos primitivos (`string`, `int`, `float64`, `bool`) y `structs`.

**Recomendación de Diseño:**
Usa structs para agrupar parámetros si son muchos. Nexus aplanará la estructura en el JSON de entrada, pero el código Go generado será más limpio.

### C. Nombres de Parámetros
Nexus normaliza los nombres para permitir flexibilidad (Fuzzy Matching).
*   En Go: `userID`
*   En JSON/Consumer: `user_id`, `userid`, `UserID` -> **Todos funcionan**.

## 4. ¿Qué librerías necesitan Refactorización?

Si tienes una librería monolítica (e.g., `mi-lib-legacy` con 50 archivos en la raíz):

1.  **Nivel Bajo de Esfuerzo**:
    *   Agregar un `lib_config.json` en la raíz con `"is_domain": true`.
    *   Resultado: API plana `client.MiLib.Funcion1`, `client.MiLib.Funcion2`.

2.  **Nivel Alto de Esfuerzo (Recomendado)**:
    *   Mover archivos a carpetas temáticas (`/users`, `/accounts`).
    *   Agregar `lib_config.json` en cada carpeta.
    *   Resultado: API estructurada `client.MiLib.Users.Create`, `client.MiLib.Accounts.Get`.

## Resumen del Contrato

1.  **Usa `lib_config.json`** para guiar al explorador de Nexus.
2.  **Organiza en carpetas** para crear una API jerárquica y limpia.
3.  **Exporta (Mayúscula)** solo lo que quieras que sea público.
