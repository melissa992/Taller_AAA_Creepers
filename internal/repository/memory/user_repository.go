package memory

import (
	"context"
	"strings"
	"sync"

	"automarket/internal/domain"
)

// UserRepository implementa domain.UserRepository utilizando mapas en memoria y sync.RWMutex.
type UserRepository struct {
	mu         sync.RWMutex
	users      map[string]*domain.User
	byUsername map[string]string
	byEmail    map[string]string
}

// NewUserRepository inicializa un repositorio de usuarios concurrente en memoria.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:      make(map[string]*domain.User),
		byUsername: make(map[string]string),
		byEmail:    make(map[string]string),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	normUsername := strings.ToLower(strings.TrimSpace(user.Username))
	normEmail := strings.ToLower(strings.TrimSpace(user.Email))

	if _, exists := r.byUsername[normUsername]; exists {
		return domain.ErrUserAlreadyExists
	}
	if _, exists := r.byEmail[normEmail]; exists {
		return domain.ErrUserAlreadyExists
	}

	// Almacenar copia defensiva para evitar modificaciones externas
	storedUser := user.Clone()
	r.users[storedUser.ID] = storedUser
	r.byUsername[normUsername] = storedUser.ID
	r.byEmail[normEmail] = storedUser.ID

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user.Clone(), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(username))
	id, exists := r.byUsername[norm]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user.Clone(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(email))
	id, exists := r.byEmail[norm]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user.Clone(), nil
}

func (r *UserRepository) ListAll(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		result = append(result, user.Clone())
	}
	return result, nil
}
