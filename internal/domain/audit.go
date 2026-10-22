package domain

import "time"

// AuditLog representa un registro de auditoría y trazabilidad (Accounting en AAA).
type AuditLog struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	UserID    string        `json:"user_id"`
	Username  string        `json:"username"`
	Role      string        `json:"role"`
	Status    int           `json:"status"`
	Latency   time.Duration `json:"latency"`
	ClientIP  string        `json:"client_ip"`
	UserAgent string        `json:"user_agent"`
}

// Clone produce una copia profunda para seguridad concurrente.
func (a *AuditLog) Clone() *AuditLog {
	if a == nil {
		return nil
	}
	copyAudit := *a
	return &copyAudit
}
