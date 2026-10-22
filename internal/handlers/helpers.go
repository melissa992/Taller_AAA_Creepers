package handlers

import (
	"strings"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

// GetCommonContext ensambla los datos comunes de sesión, navegación y alertas para las plantillas SSR.
func GetCommonContext(c *gin.Context, title string) gin.H {
	isAuth, _ := c.Get("is_authenticated")
	authenticated, _ := isAuth.(bool)

	userID := c.GetString("user_id")
	username := c.GetString("username")
	roleStr := c.GetString("role")

	data := gin.H{
		"Title":           title,
		"IsAuth":          authenticated,
		"UserID":          userID,
		"Username":        username,
		"Role":            roleStr,
		"IsAdmin":         roleStr == string(domain.RoleAdmin),
		"IsSeller":        roleStr == string(domain.RoleSeller),
		"IsBuyer":         roleStr == string(domain.RoleBuyer),
		"CurrentPath":     c.Request.URL.Path,
		"FlashError":      translateMessage(c.Query("error")),
		"FlashSuccess":    translateMessage(c.Query("success")),
		"FlashWarning":    translateMessage(c.Query("warning")),
	}

	return data
}

// translateMessage traduce códigos de mensajes a textos legibles en español.
func translateMessage(code string) string {
	if code == "" {
		return ""
	}
	switch strings.ToLower(code) {
	case "requiere_autenticacion":
		return "Debe iniciar sesión para acceder al recurso solicitado."
	case "credenciales_invalidas":
		return "Usuario o contraseña incorrectos."
	case "cuenta_creada":
		return "¡Cuenta creada exitosamente! Por favor, inicie sesión."
	case "sesion_cerrada":
		return "Ha cerrado sesión correctamente."
	case "vehiculo_creado":
		return "Vehículo publicado con éxito en el catálogo."
	case "vehiculo_actualizado":
		return "Información del vehículo actualizada correctamente."
	case "vehiculo_eliminado":
		return "El vehículo ha sido eliminado del catálogo."
	case "acceso_denegado":
		return "No tiene permisos para realizar esta operación."
	case "campos_invalidos":
		return "Por favor verifique los campos del formulario."
	default:
		return code
	}
}
