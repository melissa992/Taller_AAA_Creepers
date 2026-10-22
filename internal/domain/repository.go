package domain

import "context"

// UserRepository define los métodos para la persistencia y consulta de usuarios.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	ListAllUsers(ctx context.Context) ([]*User, error)
}

// SessionRepository define los métodos para la gestión de sesiones en memoria.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
}

// VehicleRepository define los métodos para la consulta y persistencia de vehículos.
type VehicleRepository interface {
	Create(ctx context.Context, vehicle *Vehicle) error
	GetByID(ctx context.Context, id string) (*Vehicle, error)
	GetPublished(ctx context.Context) ([]*Vehicle, error)
	GetBySellerID(ctx context.Context, sellerID string) ([]*Vehicle, error)
	ListAll(ctx context.Context) ([]*Vehicle, error)
	Update(ctx context.Context, vehicle *Vehicle) error
	Delete(ctx context.Context, id string) error
}

// AuditRepository define los métodos para el registro y consulta de eventos AAA (Accounting).
type AuditRepository interface {
	Record(ctx context.Context, entry *AuditLog) error
	ListRecent(ctx context.Context, limit int) ([]*AuditLog, error)
}
