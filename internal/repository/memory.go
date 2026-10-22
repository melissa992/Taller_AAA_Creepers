package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"
)

// MemoryRepository encapsula las colecciones en memoria con protección concurrente (sync.RWMutex).
type MemoryRepository struct {
	userMu     sync.RWMutex
	users      map[string]*domain.User
	byUsername map[string]string

	vehicleMu sync.RWMutex
	vehicles  map[string]*domain.Vehicle

	sessionMu sync.RWMutex
	sessions  map[string]*domain.Session

	auditMu sync.RWMutex
	logs    []*domain.AuditLog
	maxLogs int
}

var (
	instance *MemoryRepository
	once     sync.Once
)

// GetInstance implementa el patrón Singleton garantizando una única fuente de verdad en memoria.
func GetInstance() *MemoryRepository {
	once.Do(func() {
		instance = &MemoryRepository{
			users:      make(map[string]*domain.User),
			byUsername: make(map[string]string),
			vehicles:   make(map[string]*domain.Vehicle),
			sessions:   make(map[string]*domain.Session),
			logs:       make([]*domain.AuditLog, 0, 1000),
			maxLogs:    1000,
		}
		instance.seedData()
	})
	return instance
}

// seedData inicializa los datos de prueba (Mock Data) con vehículos en estado Published
// apuntando a "/assets/images/car1.jpg" según la HU-VIS-01.
func (r *MemoryRepository) seedData() {
	// Usuarios iniciales
	adminUser := &domain.User{
		ID:           "usr-admin",
		Username:     "admin",
		PasswordHash: security.HashPassword("Admin123!", "salt_admin_demo"),
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now(),
	}
	r.users[adminUser.ID] = adminUser
	r.byUsername["admin"] = adminUser.ID

	sellerUser := &domain.User{
		ID:           "usr-seller-1",
		Username:     "carlos_autos",
		PasswordHash: security.HashPassword("Seller123!", "salt_seller_demo"),
		Role:         domain.RoleSeller,
		CreatedAt:    time.Now(),
	}
	r.users[sellerUser.ID] = sellerUser
	r.byUsername["carlos_autos"] = sellerUser.ID

	// Vehículos Semilla (Mock Data) según la especificación exacta
	seedVehicles := []*domain.Vehicle{
		{
			ID:             "veh-001",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "Toyota Land Cruiser 300 GR-Sport 2024",
			PriceFormatted: "$135,000 USD",
			YearMileage:    "2024 • 8,500 km",
			Location:       "Bogotá, Colombia",
			WhatsAppNumber: "573001234567",
			Status:         domain.StatusPublished,
			SellerID:       sellerUser.ID,
		},
		{
			ID:             "veh-002",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "BMW M3 Competition xDrive 2023",
			PriceFormatted: "$98,500 USD",
			YearMileage:    "2023 • 14,200 km",
			Location:       "Medellín, Colombia",
			WhatsAppNumber: "573109876543",
			Status:         domain.StatusPublished,
			SellerID:       sellerUser.ID,
		},
		{
			ID:             "veh-003",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "Porsche 911 GT3 Touring Package",
			PriceFormatted: "$225,000 USD",
			YearMileage:    "2024 • 3,100 km",
			Location:       "Cali, Colombia",
			WhatsAppNumber: "573155558899",
			Status:         domain.StatusPublished,
			SellerID:       sellerUser.ID,
		},
		{
			ID:             "veh-004",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "Ford Mustang Dark Horse 5.0L V8",
			PriceFormatted: "$68,900 USD",
			YearMileage:    "2024 • 5,400 km",
			Location:       "Barranquilla, Colombia",
			WhatsAppNumber: "573204441122",
			Status:         domain.StatusPublished,
			SellerID:       sellerUser.ID,
		},
		{
			ID:             "veh-005",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "Audi RS6 Avant Performance 2023",
			PriceFormatted: "$142,000 USD",
			YearMileage:    "2023 • 19,000 km",
			Location:       "Bogotá, Colombia",
			WhatsAppNumber: "573017779900",
			Status:         domain.StatusPending, // Estado pendiente: no debe verse en catálogo público
			SellerID:       sellerUser.ID,
		},
		{
			ID:             "veh-006",
			ImageURL:       "/assets/images/car1.jpg",
			Title:          "Mercedes-AMG C63 S E-Performance",
			PriceFormatted: "$115,000 USD",
			YearMileage:    "2023 • 25,000 km",
			Location:       "Bucaramanga, Colombia",
			WhatsAppNumber: "573183332211",
			Status:         domain.StatusSold, // Estado vendido: no debe verse en catálogo público
			SellerID:       sellerUser.ID,
		},
	}

	for _, v := range seedVehicles {
		r.vehicles[v.ID] = v
	}
}

// =========================================================================
// MÉTODOS DE VEHÍCULOS (domain.VehicleRepository)
// =========================================================================

func (r *MemoryRepository) GetPublished(ctx context.Context) ([]*domain.Vehicle, error) {
	r.vehicleMu.RLock()
	defer r.vehicleMu.RUnlock()

	var result []*domain.Vehicle
	for _, v := range r.vehicles {
		if v.Status == domain.StatusPublished {
			result = append(result, v.Clone())
		}
	}
	return result, nil
}

func (r *MemoryRepository) ListAll(ctx context.Context) ([]*domain.Vehicle, error) {
	r.vehicleMu.RLock()
	defer r.vehicleMu.RUnlock()

	var result []*domain.Vehicle
	for _, v := range r.vehicles {
		result = append(result, v.Clone())
	}
	return result, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	r.vehicleMu.RLock()
	defer r.vehicleMu.RUnlock()

	v, exists := r.vehicles[id]
	if !exists {
		return nil, domain.ErrVehicleNotFound
	}
	return v.Clone(), nil
}

func (r *MemoryRepository) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	if vehicle == nil || vehicle.ID == "" {
		return domain.ErrInvalidInput
	}
	r.vehicleMu.Lock()
	defer r.vehicleMu.Unlock()

	r.vehicles[vehicle.ID] = vehicle.Clone()
	return nil
}

func (r *MemoryRepository) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	if vehicle == nil || vehicle.ID == "" {
		return domain.ErrInvalidInput
	}
	r.vehicleMu.Lock()
	defer r.vehicleMu.Unlock()

	if _, exists := r.vehicles[vehicle.ID]; !exists {
		return domain.ErrVehicleNotFound
	}
	r.vehicles[vehicle.ID] = vehicle.Clone()
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.vehicleMu.Lock()
	defer r.vehicleMu.Unlock()

	if _, exists := r.vehicles[id]; !exists {
		return domain.ErrVehicleNotFound
	}
	delete(r.vehicles, id)
	return nil
}

// =========================================================================
// MÉTODOS DE USUARIOS (domain.UserRepository)
// =========================================================================

func (r *MemoryRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidInput
	}
	r.userMu.Lock()
	defer r.userMu.Unlock()

	norm := strings.ToLower(strings.TrimSpace(user.Username))
	if _, exists := r.byUsername[norm]; exists {
		return domain.ErrUserAlreadyExists
	}

	u := user.Clone()
	r.users[u.ID] = u
	r.byUsername[norm] = u.ID
	return nil
}

func (r *MemoryRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	r.userMu.RLock()
	defer r.userMu.RUnlock()

	u, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return u.Clone(), nil
}

func (r *MemoryRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.userMu.RLock()
	defer r.userMu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(username))
	id, exists := r.byUsername[norm]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return r.users[id].Clone(), nil
}

func (r *MemoryRepository) ListAllUsers(ctx context.Context) ([]*domain.User, error) {
	r.userMu.RLock()
	defer r.userMu.RUnlock()

	var result []*domain.User
	for _, u := range r.users {
		result = append(result, u.Clone())
	}
	return result, nil
}

// =========================================================================
// MÉTODOS DE AUDITORÍA - ACCOUNTING (domain.AuditRepository)
// =========================================================================

func (r *MemoryRepository) Record(ctx context.Context, entry *domain.AuditLog) error {
	if entry == nil {
		return domain.ErrInvalidInput
	}
	r.auditMu.Lock()
	defer r.auditMu.Unlock()

	if len(r.logs) >= r.maxLogs {
		r.logs = r.logs[1:]
	}
	r.logs = append(r.logs, entry.Clone())
	return nil
}

func (r *MemoryRepository) ListRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	r.auditMu.RLock()
	defer r.auditMu.RUnlock()

	total := len(r.logs)
	if total == 0 {
		return []*domain.AuditLog{}, nil
	}
	if limit <= 0 || limit > total {
		limit = total
	}

	result := make([]*domain.AuditLog, 0, limit)
	for i := total - 1; i >= total-limit; i-- {
		result = append(result, r.logs[i].Clone())
	}
	return result, nil
}
