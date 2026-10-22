package middleware

import (
	"net/http"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session_token"
	SessionDuration   = 24 * 60 * 60 // 24 horas en segundos
)

// Authenticate verifica la presencia de la cookie de sesión y busca la sesión en el repositorio en memoria.
// Inyecta el ID del usuario en el gin.Context ("user_id", "username", "role").
func Authenticate(sessionRepo domain.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil || token == "" {
			c.Set("is_authenticated", false)
			c.Next()
			return
		}

		session, err := sessionRepo.GetSessionByToken(c.Request.Context(), token)
		if err != nil || session == nil || session.IsExpired() {
			ClearSessionCookie(c)
			c.Set("is_authenticated", false)
			c.Next()
			return
		}

		// Inyección del ID del usuario y datos de sesión en gin.Context
		c.Set("is_authenticated", true)
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Set("role", string(session.Role))
		c.Set("session", session)

		c.Next()
	}
}

// RequireAuth restringe el acceso exclusivamente a usuarios autenticados.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			SetAuditMetadata(c, "ACCESO_DENEGADO", c.Request.URL.Path, "DENEGADO_NO_AUTENTICADO")
			c.Redirect(http.StatusSeeOther, "/login?error=Debe+iniciar+sesión+para+continuar")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRoles verifica que el rol inyectado en gin.Context coincida con los roles autorizados (Authz).
func RequireRoles(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			SetAuditMetadata(c, "ACCESO_DENEGADO", c.Request.URL.Path, "DENEGADO_NO_AUTENTICADO")
			c.Redirect(http.StatusSeeOther, "/login?error=Debe+iniciar+sesión")
			c.Abort()
			return
		}

		userRoleStr := c.GetString("role")
		userRole := domain.Role(userRoleStr)

		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			SetAuditMetadata(c, "ACCESO_DENEGADO", c.Request.URL.Path, "DENEGADO_ROL_INSUFICIENTE")
			c.String(http.StatusForbidden, "403 Acceso Denegado: Se requiere rol autorizado (%s)", formatRoles(allowedRoles))
			c.Abort()
			return
		}

		c.Next()
	}
}

// SetSessionCookie configura una cookie segura con HttpOnly y SameSite.
func SetSessionCookie(c *gin.Context, token string) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
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

// ClearSessionCookie elimina la cookie en el navegador.
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

func formatRoles(roles []domain.Role) string {
	res := ""
	for i, r := range roles {
		if i > 0 {
			res += ", "
		}
		res += string(r)
	}
	return res
}
