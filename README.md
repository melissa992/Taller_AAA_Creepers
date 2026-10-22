# AutoMarket - Sistema Web Seguro con Arquitectura AAA

Sistema web para compra y venta de vehículos desarrollado en **Go puro** utilizando arquitectura Cliente-Servidor (SSR con plantillas HTML nativas) y el framework **Gin**.

Proyecto desarrollado bajo los lineamientos del taller de **Ciberseguridad: Arquitectura AAA (Authentication, Authorization, Accounting)** por el grupo **Creepers**.

---

## 🚀 Historias de Usuario Implementadas

### 1. HU-VIS-01: Catálogo Público de Vehículos (Visitante)
- **Ruta pública `GET '/'`**: Consulta en el repositorio en memoria únicamente los vehículos con estado `Published`.
- **Estructura tipo Marketplace**:
  - Imagen en la parte superior (`/assets/images/car1.jpg`).
  - Textos estructurados y claros: Título, Precio Formateado, Año / Kilometraje y Ubicación.
  - Botón de contacto directo hacia WhatsApp con enlace seguro (`wa.me`) y mensaje predeterminado.
- **Frontend sin dependencias**: Plantilla `templates/index.html` desarrollada con **CSS Grid nativo**, modular y con bajo acoplamiento de estilos (0 frameworks externos, 0 JavaScript).

### 2. HU-VEN-01: Gestión de Publicaciones del Vendedor
- **Registro de Vendedores (`POST '/register'` y `GET '/register'`):**
  - Validación robusta de inputs (campos no vacíos, longitud mínima de usuario y contraseña, coincidencia de confirmación).
  - Almacenamiento seguro de contraseñas hasheadas exclusivamente con **`crypto/sha256`** de la biblioteca estándar de Go.
  - Asignación obligatoria del rol **`Seller`**.
- **Autenticación con Sesiones en Memoria (`POST '/login'`, `GET '/login'` y `POST '/logout'`):**
  - Validación en tiempo constante contra *timing attacks* (`crypto/subtle.ConstantTimeCompare`).
  - Generación de token criptográfico seguro con `crypto/rand` (32 bytes / 64 caracteres hex).
  - Emisión de cookie segura `session_token` con directivas `HttpOnly` y `SameSite=Lax`.
- **Panel del Vendedor (`GET '/dashboard'`):**
  - Vista `templates/dashboard_seller.html` que incluye el formulario de publicación a la izquierda y el listado de vehículos propios a la derecha, con badges de estado (`Pending`, `Published`, `Sold`).
- **Publicación Protegida (`POST '/vehicles'`):**
  - Protegido por los middlewares de **Authentication** y **Authorization** (exclusivo para rol `Seller`).
  - Extracción confiable del `SellerID` a partir del `gin.Context` (sin confiar en datos enviados por el cliente).
  - Sanitización estricta de entradas (`html.EscapeString`) contra ataques de inyección XSS.
  - Almacenamiento del nuevo vehículo en estado inicial obligatorio: **`Pending`** (pendiente de aprobación).
  - Manejo de errores semántico: Si falla una validación, se re-renderiza el formulario inyectando el mensaje de error y preservando los campos válidos ingresados.

---

## 🏛️ Estructura del Proyecto

El proyecto sigue una arquitectura limpia basada en separación de responsabilidades e inyección de dependencias:

```
Taller_AAA_Creepers/
├── assets/
│   └── images/
│       └── car1.jpg                # Imagen local para los vehículos semilla del catálogo
├── cmd/
│   └── server/
│       └── main.go                 # Punto de entrada, configuración de rutas e inyección
├── internal/
│   ├── domain/                     # Entidades puras e interfaces
│   │   ├── user.go                 # Entidad User y roles (Visitor, Seller, Admin)
│   │   ├── vehicle.go              # Entidad Vehicle y estados (Pending, Published, Sold)
│   │   ├── session.go              # Entidad Session para tokens en memoria
│   │   ├── audit.go                # Entidad AuditLog para trazabilidad AAA
│   │   ├── errors.go               # Errores centinela del dominio
│   │   └── repository.go           # Interfaces de repositorios (Vehicle, User, Session, Audit)
│   ├── repository/
│   │   ├── memory.go               # Repositorio en memoria Singleton thread-safe (sync.RWMutex)
│   │   └── memory_test.go          # Pruebas de concurrencia y estrés (50 goroutines)
│   ├── security/                   # Criptografía pura con biblioteca estándar
│   │   ├── hasher.go               # Hashing con crypto/sha256 y tokens con crypto/rand
│   │   └── hasher_test.go          # Pruebas unitarias de derivación y comparación
│   ├── middleware/                 # Implementación estricta de AAA (RF-01)
│   │   ├── accounting.go           # Accounting: Log inmutable de peticiones (Método, URL, UserID, Timestamp)
│   │   └── auth.go                 # Authentication (cookie -> gin.Context) y Authorization (RequireRoles)
│   └── handlers/                   # Controladores con dependencias inyectadas
│       ├── catalog_handler.go      # Catálogo público (HU-VIS-01)
│       ├── auth_handler.go         # Registro, Login y Logout (HU-VEN-01)
│       ├── vehicle_handler.go      # Dashboard del vendedor y creación en estado Pending (HU-VEN-01)
│       └── hu_ven_01_test.go       # Suite de pruebas de integración de la HU-VEN-01
├── templates/                      # Plantillas HTML nativas (SSR)
│   ├── index.html                  # Marketplace público con CSS Grid nativo
│   ├── login.html                  # Formulario de inicio de sesión
│   ├── register.html               # Formulario de registro con rol Seller
│   └── dashboard_seller.html       # Panel del vendedor con formulario y lista de publicaciones
├── go.mod                          # ÚNICA dependencia externa permitida: gin-gonic/gin
├── go.sum
└── README.md
```

---

## 🔒 Cumplimiento de Restricciones Técnicas

| Requisito | Estado | Detalle de Implementación |
|---|---|---|
| **Lenguaje** | ✅ Go puro | Sin CGO, sin scripts externos |
| **Única Dependencia Permitida** | ✅ `github.com/gin-gonic/gin` | Cero librerías adicionales |
| **Prohibido ORMs / Base de Datos Externa** | ✅ En memoria | Mapas en memoria gestionados mediante el patrón Singleton |
| **Concurrencia Segura (Thread-Safety)** | ✅ `sync.RWMutex` | Lecturas con `RLock()` y escrituras con `Lock()`, retornando clones defensivos |
| **Prohibido JWT, bcrypt, Casbin** | ✅ Standard Library | Sesiones opacas en memoria con `crypto/rand` y hash con `crypto/sha256` |
| **Sin Variables Globales en Handlers** | ✅ Inyección de Dependencias | Repositorios pasados mediante constructores `New...Handler(repo)` |
| **Manejo de Errores y Códigos HTTP** | ✅ Semántico | Verificación explícita de `err` y respuestas con 200, 303, 400, 401, 403, 500 |
| **Frontend sin Frameworks JS/CSS** | ✅ Nativo | HTML5 semántico y CSS Grid puro embebido en las plantillas |

---

## 🛡️ Especificación del Modelo AAA (RF-01)

### 1. Authentication (Autenticación)
- Las contraseñas se almacenan como hashes hexadecimales generados por `crypto/sha256`.
- Las sesiones utilizan identificadores opacos generados con 256 bits de entropía pseudoaleatoria segura (`crypto/rand`).
- La cookie de sesión `session_token` se envía con `HttpOnly: true` (prevención de ataques XSS) y `SameSite: Lax` (prevención de CSRF).
- El middleware `Authenticate` valida la cookie, localiza la sesión en memoria y propaga al `gin.Context` los valores `"user_id"`, `"username"`, `"role"` e `"is_authenticated"`.

### 2. Authorization (Autorización)
- Roles admitidos en el sistema: `Visitor`, `Seller`, `Admin`.
- El middleware `RequireRoles(domain.RoleSeller)` valida el rol presente en el contexto; si el cliente no posee privilegios suficientes, aborta la petición con `403 Forbidden`.
- Las rutas privadas de publicación y consulta del vendedor (`/dashboard`, `/vehicles`) están protegidas contra accesos no autorizados.

### 3. Accounting (Auditoría y Trazabilidad)
- Middleware global `middleware.Accounting(repo)` activo para **todas** las rutas del sistema.
- Registra de forma estructurada en consola y en memoria:
  - **Método HTTP** (`GET`, `POST`, etc.)
  - **URL / Endpoint solicitado**
  - **UserID y Username** (`ANONYMOUS` si no hay sesión activa)
  - **Rol asignado**
  - **Código de Estado HTTP** (`200`, `303`, `401`, `403`, etc.)
  - **Latencia de procesamiento**
  - **Timestamp en formato UTC**

---

## 👥 Cuentas Mock Semilla para Pruebas

El repositorio incluye datos precargados para validación inmediata:

| Usuario | Contraseña | Rol | Acceso |
|---|---|---|---|
| **`carlos_autos`** | `Seller123!` | `Seller` | Acceso al panel `/dashboard` y publicación de vehículos |
| **`admin`** | `Admin123!` | `Admin` | Rol administrativo |

*(También es posible registrar cualquier usuario nuevo desde `/register`, el cual obtendrá automáticamente el rol `Seller`).*

---

## 🚀 Instrucciones de Ejecución

### 1. Prerrequisitos
- Go 1.24+ instalado en el entorno.

### 2. Iniciar el Servidor
```powershell
go run ./cmd/server
```

La aplicación quedará disponible en:  
👉 **`http://localhost:8080/`**

### 3. Rutas Disponibles
- **Catálogo Público (HU-VIS-01):** `http://localhost:8080/`
- **Registro de Vendedores (HU-VEN-01):** `http://localhost:8080/register`
- **Inicio de Sesión:** `http://localhost:8080/login`
- **Panel del Vendedor (Protegido):** `http://localhost:8080/dashboard`

---

## 🧪 Ejecución de Pruebas Automatizadas

El proyecto cuenta con suites de pruebas unitarias y de integración:

```powershell
go test -v ./...
```

### Cobertura de Pruebas:
1. `internal/repository`: Valida el Singleton, los datos semilla en estado `Published` y la concurrencia segura con 50 goroutines bajo `sync.RWMutex`.
2. `internal/security`: Valida el hasheo SHA-256, la comparación en tiempo constante y la generación de tokens.
3. `internal/handlers`: Valida el flujo completo de la HU-VEN-01 (registro, login con cookie, rechazo a no autenticados y publicación de vehículo con estado `Pending`).
