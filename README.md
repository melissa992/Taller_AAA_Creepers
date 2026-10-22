# AutoMarket - Sistema Web Seguro con Arquitectura AAA

Sistema web para compra y venta de vehículos desarrollado en **Go puro** utilizando arquitectura Cliente-Servidor (SSR con plantillas HTML nativas) y el framework **Gin**.

Diseñado bajo estándares de ciberseguridad para la gestión segura de accesos mediante el modelo **AAA (Authentication, Authorization, Accounting)** según los lineamientos del taller (Grupo Creepers).

---

## 🏛️ Arquitectura del Sistema

El proyecto sigue una separación rigurosa de responsabilidades con alta cohesión y bajo acoplamiento:

```
Taller_AAA_Creepers/
├── cmd/
│   └── server/
│       └── main.go                 # Punto de entrada, bootstrap e inyección de dependencias
├── internal/
│   ├── domain/                     # Entidades puras e interfaces de repositorio
│   │   ├── user.go                 # Entidad User y roles (ADMIN, SELLER, BUYER)
│   │   ├── vehicle.go              # Entidad Vehicle y filtros de búsqueda
│   │   ├── session.go              # Entidad Session para tokens en memoria
│   │   ├── audit.go                # Entidad AuditLog para trazabilidad AAA
│   │   ├── errors.go               # Errores centinela del dominio
│   │   └── repository.go           # Interfaces de los repositorios
│   ├── repository/
│   │   └── memory/                 # Implementación en memoria con sync.RWMutex
│   │       ├── user_repository.go
│   │       ├── session_repository.go
│   │       ├── vehicle_repository.go
│   │       ├── audit_repository.go
│   │       └── repository_test.go  # Pruebas de concurrencia y estrés
│   ├── security/                   # Criptografía pura con paquetes estándar
│   │   ├── hasher.go               # Hashing con crypto/sha256, sal y token con crypto/rand
│   │   └── hasher_test.go          # Pruebas de verificación de contraseñas y tokens
│   ├── middleware/                 # Implementación estricta de AAA (RF-01)
│   │   ├── auth.go                 # Authentication: Cookie de sesión HttpOnly -> Memoria
│   │   ├── rbac.go                 # Authorization: Verificación de roles inyectados en Gin
│   │   ├── accounting.go           # Accounting: Intercepción, logging y auditoría inmutable
│   │   └── aaa_test.go             # Pruebas de integración del flujo AAA
│   ├── handlers/                   # Controladores Gin con interfaces inyectadas
│   │   ├── auth_handler.go         # Login, Registro, Logout con cookies seguras
│   │   ├── vehicle_handler.go      # Catálogo, Detalle, CRUD y verificación de propiedad
│   │   ├── dashboard_handler.go    # Dashboard por rol y visor de Auditoría AAA
│   │   └── helpers.go              # Contexto unificado y mensajes de vista
│   └── bootstrap/
│       └── seed.go                 # Carga inicial de usuarios y catálogo de vehículos
├── web/
│   ├── static/
│   │   └── css/
│   │       └── style.css           # Estilos modernos con paleta cuidada (sin Tailwind)
│   └── templates/                  # Vistas Server-Side Rendering (SSR)
│       ├── layouts/                # Base, Navbar con estado de rol, Flash y Footer
│       ├── auth/                   # Formularios de Login y Registro
│       ├── vehicles/               # Catálogo, Detalle, Crear y Editar
│       ├── dashboard/              # Panel general y Visor de Auditoría AAA
│       └── errors/                 # Vistas de error semánticas (403, 404, 500)
├── go.mod                          # Módulo Go (Única dependencia: gin-gonic/gin)
└── go.sum
```

---

## 🔒 Cumplimiento Estricto de Restricciones Técnicas

| Requisito | Cumplimiento | Detalle Técnico |
|---|---|---|
| **Lenguaje** | ✅ Go Puro | Sin CGO, sin scripts externos |
| **Única Dependencia** | ✅ `github.com/gin-gonic/gin` | Cero librerías adicionales |
| **Prohibido ORMs / DB** | ✅ Todo en memoria | Tablas virtuales implementadas como mapas Go |
| **Concurrencia Segura** | ✅ `sync.RWMutex` | Lecturas concurrentes con `RLock()` y escrituras con `Lock()` |
| **Prohibido JWT / bcrypt** | ✅ Standard Library | Sesiones opacas con `crypto/rand` y hash con `crypto/sha256` |
| **Sin Variables Globales** | ✅ Inyección de Dependencias | Todos los Handlers reciben las interfaces del dominio |
| **Manejo de Errores** | ✅ Errores Verificados | Cero `err` ignorados; códigos HTTP semánticos (200, 303, 400, 401, 403, 404, 500) |

---

## 🛡️ RF-01: Gestión Segura AAA (Detalle de Implementación)

### 1. Authentication (Autenticación)
- **Almacenamiento de Contraseñas:** 
  - Sal criptográfica aleatoria de 16 bytes generada con `crypto/rand`.
  - Derivación mediante `crypto/sha256.Sum256` concatenando `salt:password`.
  - Verificación contra ataques de canal lateral (Timing Attacks) usando `crypto/subtle.ConstantTimeCompare`.
- **Manejo de Sesiones:**
  - Token aleatorio de 32 bytes (64 caracteres hex) con alta entropía criptográfica.
  - Almacenamiento en `SessionRepository` con tiempo de expiración (24 horas).
  - Cookie HTTP configurada con directivas de seguridad:
    - `HttpOnly: true` (inmune a robo de sesión por XSS en JavaScript).
    - `SameSite: http.SameSiteLaxMode` (protección contra ataques CSRF).
    - `Secure: true` adaptativo (automático si se detecta conexión TLS/HTTPS).

### 2. Authorization (Autorización)
- **Control de Acceso Basado en Roles (RBAC):**
  - Roles soportados: `ADMIN`, `SELLER` (Vendedor), `BUYER` (Comprador).
  - Middleware `RequireRoles(roles ...domain.Role)`:
    - Extrae el rol validado previamente en el contexto de Gin.
    - Si el usuario no tiene privilegios suficientes, emite un código HTTP `403 Forbidden` semántico y renderiza una vista explicativa con el rol actual y los roles requeridos.
- **Autorización a Nivel de Recurso (Ownership Control):**
  - Al modificar o eliminar un vehículo, el controlador verifica que `vehicle.SellerID == currentUserID` o que el usuario posea rol `ADMIN`.

### 3. Accounting (Auditoría y Trazabilidad)
- **Middleware Global `middleware.Accounting(auditRepo)`:**
  - Intercepta **todas** las peticiones HTTP entrantes.
  - Registra de forma estructurada e inmutable:
    - **Método HTTP** (`GET`, `POST`, `PUT`, `DELETE`).
    - **URL / Ruta del Recurso** (`/vehicles`, `/login`, etc.).
    - **UserID y Username** (o `ANONYMOUS` si no hay sesión activa).
    - **Rol del Usuario** (`ADMIN`, `SELLER`, `BUYER`, `NONE`).
    - **Timestamp UTC** del momento exacto del evento.
    - **Código de Estado HTTP** resultante (`200`, `303`, `401`, `403`, etc.).
    - **Latencia de Ejecución**.
    - **Dirección IP del Cliente** y **User-Agent**.
  - **Doble salida:**
    1. Salida estándar de consola con formato estructurado `[ACCOUNTING] ...`.
    2. Almacenamiento concurrente en `AuditRepository` para consulta gráfica en la vista web `/dashboard/audit` (exclusiva para administradores).

---

## 🚀 Puesta en Marcha

### Prerrequisitos
- Go 1.24+ instalado en el sistema.

### Ejecución Directa
```bash
# Descargar dependencias (únicamente Gin)
go mod download

# Iniciar el servidor
go run ./cmd/server
```

El servidor quedará escuchando en:  
👉 **`http://localhost:8080`**

---

## 👥 Cuentas de Demostración para Pruebas

El sistema incluye datos semilla precargados con botones de auto-llenado en la pantalla de inicio de sesión:

| Usuario | Contraseña | Rol | Privilegios Principales |
|---|---|---|---|
| **`admin`** | `Admin123!` | `ADMIN` | Acceso total, eliminación de cualquier vehículo, visor de **Auditoría AAA (Accounting)** y lista de usuarios |
| **`vendedor1`** | `Vendedor123!` | `SELLER` | Publicar nuevos vehículos, editar y eliminar exclusivamente **sus propios** vehículos |
| **`comprador1`** | `Comprador123!` | `BUYER` | Explorar catálogo con filtros, ver fichas técnicas completas; bloqueado para publicar/editar (403 Forbidden) |

---

## 🧪 Ejecución de Pruebas Automatizadas

El proyecto cuenta con suites de pruebas unitarias y de integración que validan:
1. Concurrencia segura de lectura/escritura (`sync.RWMutex`) en repositorios.
2. Generación segura de hashes SHA-256 y tokens de sesión.
3. Flujo completo del middleware AAA (Autenticación, Autorización y Accounting).

Para correr las pruebas:
```bash
go test -v ./internal/...
```
