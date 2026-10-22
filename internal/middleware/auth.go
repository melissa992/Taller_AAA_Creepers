package middleware

import (
	"net/http"
	"strings"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "automarket_session"
	SessionDuration   = 24 * 60 * 60 // 24 horas en segundos
)

// Authenticate verifica la presencia de la cookie de sesión y asocia los datos del usuario al contexto.
// No bloquea la petición si no hay sesión, permitiendo que rutas públicas conozcan el estado del usuario.
func Authenticate(sessionRepo domain.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil || token == "" {
			c.Set("is_authenticated", false)
			c.Next()
			return
		}

		session, err := sessionRepo.GetByToken(c.Request.Context(), token)
		if err != nil || session == nil || session.IsExpired() {
			// Cookie inválida o expirada: eliminar del navegador
			ClearSessionCookie(c)
			c.Set("is_authenticated", false)
			c.Next()
			return
		}

		// Inyección de la identidad en el contexto de Gin
		c.Set("is_authenticated", true)
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Set("role", string(session.Role))
		c.Set("session", session)

		c.Next()
	}
}

// RequireAuth restringe el acceso exclusivamente a usuarios autenticados.
// Si el cliente no está autenticado, responde con 401 (API) o redirige a /login (SSR).
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			if isBrowserRequest(c) {
				c.Redirect(http.StatusSeeOther, "/login?error=requiere_autenticacion")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Debe autenticarse para acceder a este recurso",
			})
			return
		}
		c.Next()
	}
}

// SetSessionCookie configura una cookie de sesión con directivas de máxima seguridad.
// HttpOnly previene acceso desde scripts del navegador (mitigando ataques XSS).
// SameSite=Lax protege contra ataques de falsificación de petición en sitios cruzados (CSRF).
func SetSessionCookie(c *gin.Context, token string, secure bool) {
	// Si la conexión entrante es TLS o el flag secure está activo, marcar cookie como Secure
	isSecure := secure || c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   SessionDuration,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie invalida la cookie de sesión en el navegador.
func ClearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// isBrowserRequest determina si la petición proviene de un navegador web para vistas SSR.
func isBrowserRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
