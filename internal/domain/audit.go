package domain

import "time"

// AuditLog representa un registro de auditoría y trazabilidad (Accounting en AAA).
type AuditLog struct {
	ID        string        `json:"id"`
	Seq       int           `json:"seq"`       // Número de secuencia (1, 2, 3...)
	Timestamp time.Time     `json:"timestamp"`
	ActorID   string        `json:"actor_id"`
	Username  string        `json:"username"`  // Nombre del actor o "Visitante"
	Role      string        `json:"role"`      // Rol: Visitor, Seller, Admin
	Action    string        `json:"action"`    // Acción de negocio (ej. APROBAR_PUBLICACION)
	Resource  string        `json:"resource"`  // Recurso afectado (ej. veh-005)
	Result    string        `json:"result"`    // Resultado (EXITO, DENEGADO, ERROR)
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	Status    int           `json:"status"`
	Latency   time.Duration `json:"latency"`
	ClientIP  string        `json:"client_ip"`
	PrevHash  string        `json:"prev_hash"` 
	Hash      string        `json:"hash"`      
}

// Clone produce una copia profunda para seguridad concurrente.
func (a *AuditLog) Clone() *AuditLog {
	if a == nil {
		return nil
	}
	copyAudit := *a
	return &copyAudit
}
