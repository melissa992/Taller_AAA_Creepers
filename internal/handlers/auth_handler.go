package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"automarket/internal/domain"
	"automarket/internal/middleware"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

// AuthHandler gestiona el registro y login con inyección de dependencias.
type AuthHandler struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

// NewAuthHandler crea una nueva instancia inyectando los repositorios necesarios.
func NewAuthHandler(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// ShowRegister muestra el formulario de registro.
func (h *AuthHandler) ShowRegister(c *gin.Context) {
	if isAuth, _ := c.Get("is_authenticated"); isAuth == true {
		if c.GetString("role") == string(domain.RoleAdmin) {
			c.Redirect(http.StatusSeeOther, "/admin/publicaciones")
			return
		}
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{
		"Title": "AutoMarket - Registro de Vendedor",
	})
}

// Register valida inputs, hashea con 'crypto/sha256' y asigna el rol 'Seller'.
func (h *AuthHandler) Register(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validación robusta de inputs
	if username == "" || password == "" {
		middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", "/register", "FALLO_VALIDACION")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":        "AutoMarket - Registro de Vendedor",
			"ErrorMessage": "Todos los campos son obligatorios.",
			"Username":     username,
		})
		return
	}

	if len(username) < 3 {
		middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", username, "FALLO_VALIDACION")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":        "AutoMarket - Registro de Vendedor",
			"ErrorMessage": "El nombre de usuario debe tener al menos 3 caracteres.",
			"Username":     username,
		})
		return
	}

	if len(password) < 6 {
		middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", username, "FALLO_VALIDACION")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":        "AutoMarket - Registro de Vendedor",
			"ErrorMessage": "La contraseña debe tener al menos 6 caracteres.",
			"Username":     username,
		})
		return
	}

	if password != confirmPassword {
		middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", username, "FALLO_VALIDACION")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":        "AutoMarket - Registro de Vendedor",
			"ErrorMessage": "Las contraseñas no coinciden.",
			"Username":     username,
		})
		return
	}

	// Hasheo con 'crypto/sha256' y asignación fija de rol 'Seller' (HU-VEN-01)
	passwordHash := security.HashPassword(password)

	newUser := &domain.User{
		ID:           security.GenerateID("usr"),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         domain.RoleSeller,
		CreatedAt:    time.Now(),
	}

	if err := h.userRepo.CreateUser(c.Request.Context(), newUser); err != nil {
		errMsg := "Error interno al crear usuario."
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			errMsg = "El nombre de usuario ya está registrado."
		}
		middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", username, "FALLO_USUARIO_EXISTENTE")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":        "AutoMarket - Registro de Vendedor",
			"ErrorMessage": errMsg,
			"Username":     username,
		})
		return
	}

	c.Set("user_id", newUser.ID)
	c.Set("username", newUser.Username)
	c.Set("role", string(newUser.Role))
	middleware.SetAuditMetadata(c, "REGISTRAR_USUARIO", newUser.ID, "EXITO")

	// Redirigir al login con mensaje exitoso
	c.Redirect(http.StatusSeeOther, "/login?success=Cuenta+creada+exitosamente.+Por+favor+inicie+sesión.")
}

// ShowLogin muestra el formulario de inicio de sesión.
func (h *AuthHandler) ShowLogin(c *gin.Context) {
	if isAuth, _ := c.Get("is_authenticated"); isAuth == true {
		if c.GetString("role") == string(domain.RoleAdmin) {
			c.Redirect(http.StatusSeeOther, "/admin/publicaciones")
			return
		}
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title":        "AutoMarket - Iniciar Sesión",
		"ErrorMessage": c.Query("error"),
		"SuccessMsg":   c.Query("success"),
	})
}

// Login valida inputs, verifica credenciales y emite cookie de sesión segura.
func (h *AuthHandler) Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	if username == "" || password == "" {
		middleware.SetAuditMetadata(c, "INICIAR_SESION", "/login", "FALLO_VALIDACION")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title":        "AutoMarket - Iniciar Sesión",
			"ErrorMessage": "Por favor ingrese usuario y contraseña.",
			"Username":     username,
		})
		return
	}

	user, err := h.userRepo.GetUserByUsername(c.Request.Context(), username)
	if err != nil || !security.CheckPassword(password, user.PasswordHash) {
		c.Set("username", username)
		middleware.SetAuditMetadata(c, "INICIAR_SESION", username, "DENEGADO_CREDENCIALES_INVALIDAS")
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Title":        "AutoMarket - Iniciar Sesión",
			"ErrorMessage": "Credenciales inválidas. Verifique su usuario y contraseña.",
			"Username":     username,
		})
		return
	}

	// Generar token y registrar sesión
	token, err := security.GenerateSessionToken()
	if err != nil {
		middleware.SetAuditMetadata(c, "INICIAR_SESION", user.ID, "ERROR")
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Title":        "AutoMarket - Iniciar Sesión",
			"ErrorMessage": "Error generando token de sesión.",
		})
		return
	}

	session := &domain.Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := h.sessionRepo.CreateSession(c.Request.Context(), session); err != nil {
		middleware.SetAuditMetadata(c, "INICIAR_SESION", user.ID, "ERROR")
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Title":        "AutoMarket - Iniciar Sesión",
			"ErrorMessage": "Error almacenando la sesión.",
		})
		return
	}

	// Establecer cookie HttpOnly
	middleware.SetSessionCookie(c, token)

	c.Set("user_id", user.ID)
	c.Set("username", user.Username)
	c.Set("role", string(user.Role))
	middleware.SetAuditMetadata(c, "INICIAR_SESION", user.ID, "EXITO")

	if user.Role == domain.RoleAdmin {
		c.Redirect(http.StatusSeeOther, "/admin/publicaciones")
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Logout cierra la sesión y limpia la cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(middleware.SessionCookieName)
	if err == nil && token != "" {
		_ = h.sessionRepo.DeleteSession(c.Request.Context(), token)
	}
	middleware.ClearSessionCookie(c)
	middleware.SetAuditMetadata(c, "CERRAR_SESION", c.GetString("user_id"), "EXITO")
	c.Redirect(http.StatusSeeOther, "/login?success=Sesión+cerrada+correctamente.")
}
