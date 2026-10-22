package middleware

import (
	"log"
	"net/http"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

// Accounting implementa el componente 'Accounting' de la arquitectura AAA (RF-01).
// Registra fecha y hora, actor o visitante, acción, recurso y resultado, encadenando
// cada registro mediante SHA-256 en el repositorio y verificando si falla su escritura.
func Accounting(auditRepo domain.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		timestamp := start.UTC()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// Procesar la cadena de middlewares y handlers
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Extraer la identidad inyectada por el middleware de autenticación
		actorID := c.GetString("user_id")
		if actorID == "" {
			actorID = "VISITANTE"
		}

		username := c.GetString("username")
		if username == "" {
			username = "Visitante"
		}

		role := c.GetString("role")
		if role == "" {
			role = string(domain.RoleVisitor)
		}

		// Extraer acción, recurso y resultado definidos por los handlers o middlewares
		action := c.GetString("audit_action")
		if action == "" {
			action = method + " " + path
		}

		resource := c.GetString("audit_resource")
		if resource == "" {
			resource = path
		}

		result := c.GetString("audit_result")
		if result == "" {
			switch {
			case status >= 200 && status < 400:
				result = "EXITO"
			case status == http.StatusUnauthorized || status == http.StatusForbidden:
				result = "DENEGADO"
			default:
				result = "ERROR"
			}
		}

		entry := &domain.AuditLog{
			ID:        security.GenerateID("aud"),
			Timestamp: timestamp,
			ActorID:   actorID,
			Username:  username,
			Role:      role,
			Action:    action,
			Resource:  resource,
			Result:    result,
			Method:    method,
			URL:       path,
			Status:    status,
			Latency:   latency,
			ClientIP:  clientIP,
		}

		// Guardar en el repositorio verificando si falla la escritura (Tarea 5)
		if auditRepo != nil {
			if err := auditRepo.Record(c.Request.Context(), entry); err != nil {
				log.Printf("[ALERTA CRÍTICA AUDITORÍA] Fallo al escribir registro de auditoría: %v", err)
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "Operación abortada: no fue posible garantizar el registro de auditoría",
					})
				}
				return
			}
		}

		// Registro estructurado en salida estándar incluyendo Acción, Recurso, Resultado y Hash
		log.Printf("[ACCOUNTING] #%d %s | Actor: %s (%s) | Rol: %-7s | Acción: %-24s | Recurso: %-22s | Resultado: %-10s | Status: %d | Hash: %.12s...",
			entry.Seq,
			timestamp.Format("2006-01-02 15:04:05"),
			actorID,
			username,
			role,
			action,
			resource,
			result,
			status,
			entry.Hash,
		)
	}
}

// SetAuditMetadata permite a los handlers y middlewares especificar la acción, el recurso y el resultado de negocio.
func SetAuditMetadata(c *gin.Context, action, resource, result string) {
	c.Set("audit_action", action)
	c.Set("audit_resource", resource)
	c.Set("audit_result", result)
}
