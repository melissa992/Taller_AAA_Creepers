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

// AuthHandler gestiona las peticiones de autenticación y registro de usuarios en AutoMarket.
type AuthHandler struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

// NewAuthHandler construye el controlador con inyección de dependencias.
func NewAuthHandler(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// ShowLogin muestra el formulario de inicio de sesión.
func (h *AuthHandler) ShowLogin(c *gin.Context) {
	if isAuth, _ := c.Get("is_authenticated"); isAuth == true {
		c.Redirect(http.StatusSeeOther, "/vehicles")
		return
	}

	data := GetCommonContext(c, "Iniciar Sesión - AutoMarket")
	data["Redirect"] = c.Query("redirect")
	c.HTML(http.StatusOK, "auth/login.html", data)
}

// Login procesa las credenciales recibidas, valida con SHA-256 y crea la sesión en memoria.
func (h *AuthHandler) Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	redirectURL := c.PostForm("redirect")

	if username == "" || password == "" {
		data := GetCommonContext(c, "Iniciar Sesión - AutoMarket")
		data["ErrorMessage"] = "Debe proporcionar usuario y contraseña"
		c.HTML(http.StatusBadRequest, "auth/login.html", data)
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		data := GetCommonContext(c, "Iniciar Sesión - AutoMarket")
		data["ErrorMessage"] = "Credenciales incorrectas"
		c.HTML(http.StatusUnauthorized, "auth/login.html", data)
		return
	}

	// Verificación segura con crypto/sha256 y comparación en tiempo constante
	if !security.CheckPassword(password, user.Salt, user.PasswordHash) {
		data := GetCommonContext(c, "Iniciar Sesión - AutoMarket")
		data["ErrorMessage"] = "Credenciales incorrectas"
		c.HTML(http.StatusUnauthorized, "auth/login.html", data)
		return
	}

	// Generar token criptográficamente seguro
	token, err := security.GenerateSessionToken()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error del Servidor",
			"Error": "No se pudo generar el token de sesión",
		})
		return
	}

	session := &domain.Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(middleware.SessionDuration) * time.Second),
	}

	if err := h.sessionRepo.Create(c.Request.Context(), session); err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error del Servidor",
			"Error": "No se pudo almacenar la sesión",
		})
		return
	}

	// Configurar cookie de sesión segura (HttpOnly, SameSite)
	middleware.SetSessionCookie(c, token, false)

	target := "/vehicles"
	if redirectURL != "" && strings.HasPrefix(redirectURL, "/") {
		target = redirectURL
	}
	c.Redirect(http.StatusSeeOther, target)
}

// ShowRegister muestra el formulario de registro de nuevos usuarios.
func (h *AuthHandler) ShowRegister(c *gin.Context) {
	if isAuth, _ := c.Get("is_authenticated"); isAuth == true {
		c.Redirect(http.StatusSeeOther, "/vehicles")
		return
	}

	data := GetCommonContext(c, "Registro de Usuario - AutoMarket")
	c.HTML(http.StatusOK, "auth/register.html", data)
}

// Register valida los campos, genera la sal criptográfica, hashea la contraseña y persiste el usuario.
func (h *AuthHandler) Register(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	roleInput := strings.ToUpper(strings.TrimSpace(c.PostForm("role")))

	data := GetCommonContext(c, "Registro de Usuario - AutoMarket")

	if username == "" || email == "" || password == "" {
		data["ErrorMessage"] = "Todos los campos marcados son obligatorios"
		c.HTML(http.StatusBadRequest, "auth/register.html", data)
		return
	}

	if len(password) < 6 {
		data["ErrorMessage"] = "La contraseña debe tener al menos 6 caracteres"
		c.HTML(http.StatusBadRequest, "auth/register.html", data)
		return
	}

	if password != confirmPassword {
		data["ErrorMessage"] = "Las contraseñas no coinciden"
		c.HTML(http.StatusBadRequest, "auth/register.html", data)
		return
	}

	// Roles permitidos para auto-registro (BUYER o SELLER; ADMIN es reservado)
	var assignedRole domain.Role
	switch roleInput {
	case string(domain.RoleSeller):
		assignedRole = domain.RoleSeller
	default:
		assignedRole = domain.RoleBuyer
	}

	salt, err := security.GenerateSalt()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error Interno",
			"Error": "Fallo en la generación de seguridad criptográfica",
		})
		return
	}

	passwordHash := security.HashPassword(password, salt)

	newUser := &domain.User{
		ID:           security.GenerateID("usr"),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         assignedRole,
		CreatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(c.Request.Context(), newUser); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			data["ErrorMessage"] = "El nombre de usuario o correo ya se encuentra en uso"
			c.HTML(http.StatusConflict, "auth/register.html", data)
			return
		}
		data["ErrorMessage"] = "Error al registrar el usuario en el sistema"
		c.HTML(http.StatusInternalServerError, "auth/register.html", data)
		return
	}

	c.Redirect(http.StatusSeeOther, "/login?success=cuenta_creada")
}

// Logout destruye la sesión en memoria y remueve la cookie del cliente.
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(middleware.SessionCookieName)
	if err == nil && token != "" {
		_ = h.sessionRepo.Delete(c.Request.Context(), token)
	}

	middleware.ClearSessionCookie(c)
	c.Redirect(http.StatusSeeOther, "/login?success=sesion_cerrada")
}
