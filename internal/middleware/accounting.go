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
// y captura de forma no repudiable: Método, URL, ActorID, Acción, Recurso, Resultado y Timestamp,
// almacenándolos en memoria, archivo JSON local y en la salida estándar estructurada.
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
			actorID = "ANONYMOUS"
		}

		username := c.GetString("username")
		if username == "" {
			username = "Visitante"
		}

		role := c.GetString("role")
		if role == "" {
			role = "Visitor"
		}

		// Extraer metadatos de auditoría (si fueron definidos por middlewares/handlers) o usar valores por defecto
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
			case status >= 500:
				result = "ERROR"
			case status == 401 || status == 403:
				result = "DENEGADO"
			case status >= 400:
				result = "FALLO"
			default:
				result = "EXITO"
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

		// Registro estructurado en log estándar de Go
		log.Printf("[ACCOUNTING] %s | %-7s %s | Actor: %s (%s) | Role: %-7s | Action: %s | Result: %s | Status: %d | Latency: %v | IP: %s",
			timestamp.Format("2006-01-02 15:04:05"),
			method,
			path,
			actorID,
			username,
			role,
			action,
			result,
			status,
			latency,
			clientIP,
		)

		// Guardar en el repositorio (memoria + archivo JSON local)
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
