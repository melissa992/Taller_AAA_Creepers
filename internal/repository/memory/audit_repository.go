package memory

import (
	"context"
	"sync"

	"automarket/internal/domain"
)

// AuditRepository almacena registros de Accounting (AAA) en memoria con concurrencia segura.
type AuditRepository struct {
	mu      sync.RWMutex
	logs    []*domain.AuditLog
	maxSize int
}

// NewAuditRepository inicializa el repositorio de auditoría con un tamaño máximo para evitar desbordamiento de memoria.
func NewAuditRepository(maxSize int) *AuditRepository {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &AuditRepository{
		logs:    make([]*domain.AuditLog, 0, maxSize),
		maxSize: maxSize,
	}
}

func (r *AuditRepository) Record(ctx context.Context, entry *domain.AuditLog) error {
	if entry == nil {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Si se alcanza el límite, descartar el registro más antiguo (política FIFO)
	if len(r.logs) >= r.maxSize {
		r.logs = r.logs[1:]
	}

	r.logs = append(r.logs, entry.Clone())
	return nil
}

func (r *AuditRepository) ListRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.logs)
	if total == 0 {
		return []*domain.AuditLog{}, nil
	}

	if limit <= 0 || limit > total {
		limit = total
	}

	// Devolver en orden cronológico inverso (los más recientes primero)
	result := make([]*domain.AuditLog, 0, limit)
	for i := total - 1; i >= total-limit; i-- {
		result = append(result, r.logs[i].Clone())
	}

	return result, nil
}
