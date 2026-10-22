package domain

import "context"

// UserRepository define los métodos para la persistencia y consulta de usuarios.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListAll(ctx context.Context) ([]*User, error)
}

// SessionRepository define los métodos para la gestión de tokens de sesión en memoria.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// VehicleRepository define los métodos para el catálogo de vehículos.
type VehicleRepository interface {
	Create(ctx context.Context, vehicle *Vehicle) error
	GetByID(ctx context.Context, id string) (*Vehicle, error)
	Update(ctx context.Context, vehicle *Vehicle) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter VehicleFilter) ([]*Vehicle, error)
}

// AuditRepository define los métodos para el registro y consulta de eventos AAA (Accounting).
type AuditRepository interface {
	Record(ctx context.Context, entry *AuditLog) error
	ListRecent(ctx context.Context, limit int) ([]*AuditLog, error)
}
