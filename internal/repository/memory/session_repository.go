package memory

import (
	"context"
	"sync"
	"time"

	"automarket/internal/domain"
)

// SessionRepository implementa domain.SessionRepository para almacenar sesiones en memoria de forma segura.
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

// NewSessionRepository inicializa el repositorio de sesiones en memoria.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session == nil || session.Token == "" {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.Token] = session.Clone()
	return nil
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	r.mu.RLock()
	session, exists := r.sessions[token]
	r.mu.RUnlock()

	if !exists {
		return nil, domain.ErrSessionNotFound
	}

	if session.IsExpired() {
		// Limpieza de sesión vencida
		r.mu.Lock()
		delete(r.sessions, token)
		r.mu.Unlock()
		return nil, domain.ErrSessionExpired
	}

	return session.Clone(), nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, token)
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for token, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, token)
		}
	}
	return nil
}
