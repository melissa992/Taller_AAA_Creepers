package middleware

import (
	"log"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

// Accounting implementa el componente 'Accounting' de la arquitectura AAA (RF-01).
// Intercepta todas las peticiones entrantes, calcula la latencia, registra el estado HTTP
// y captura de forma no repudiable: Método, URL, UserID y Timestamp, almacenándolos en
// memoria y en la salida estándar estructurada.
func Accounting(auditRepo domain.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		timestamp := start.UTC()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Procesar la cadena de middlewares y handlers
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Extraer la identidad inyectada por el middleware de autenticación
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "ANONYMOUS"
		}

		username := c.GetString("username")
		if username == "" {
			username = "Invitado"
		}

		role := c.GetString("role")
		if role == "" {
			role = "NONE"
		}

		entry := &domain.AuditLog{
			ID:        security.GenerateID("aud"),
			Timestamp: timestamp,
			Method:    method,
			URL:       path,
			UserID:    userID,
			Username:  username,
			Role:      role,
			Status:    status,
			Latency:   latency,
			ClientIP:  clientIP,
			UserAgent: userAgent,
		}

		// Registro estructurado en log estándar de Go
		log.Printf("[ACCOUNTING] %s | %-7s %s | User: %s (%s) | Role: %-6s | Status: %d | Latency: %v | IP: %s",
			timestamp.Format("2006-01-02 15:04:05"),
			method,
			path,
			userID,
			username,
			role,
			status,
			latency,
			clientIP,
		)

		// Guardar en el repositorio en memoria para consulta en el Dashboard de Auditoría
		if auditRepo != nil {
			_ = auditRepo.Record(c.Request.Context(), entry)
		}
	}
}

func SetAuditMetadata(c *gin.Context, action, resource, result string) {
    c.Set("audit_action", action)
    c.Set("audit_resource", resource)
    c.Set("audit_result", result)
}