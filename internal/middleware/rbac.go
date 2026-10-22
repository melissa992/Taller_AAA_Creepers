package middleware

import (
	"net/http"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

// RequireRoles implementa el componente 'Authorization' de la arquitectura AAA (RF-01).
// Inspecciona el rol inyectado en el contexto por el middleware de autenticación
// y valida si pertenece al conjunto de roles con privilegios permitidos para la ruta.
func RequireRoles(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, exists := c.Get("is_authenticated")
		if !exists || !isAuth.(bool) {
			if isBrowserRequest(c) {
				c.Redirect(http.StatusSeeOther, "/login?error=requiere_autenticacion")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Debe autenticarse primero",
			})
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
			if isBrowserRequest(c) {
				// Renderizar vista semántica 403 Forbidden con información contextual
				c.HTML(http.StatusForbidden, "errors/403.html", gin.H{
					"Title":       "Acceso Denegado (403)",
					"UserRole":    userRoleStr,
					"Username":    c.GetString("username"),
					"Required":    formatRoles(allowedRoles),
					"IsAuth":      true,
					"CurrentPath": c.Request.URL.Path,
				})
				c.Abort()
				return
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":          "Acceso denegado: rol insuficiente para realizar esta acción",
				"current_role":   userRoleStr,
				"required_roles": allowedRoles,
			})
			return
		}

		c.Next()
	}
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
