package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"automarket/internal/domain"
)

// VehicleRepository implementa domain.VehicleRepository en memoria con concurrencia segura.
type VehicleRepository struct {
	mu       sync.RWMutex
	vehicles map[string]*domain.Vehicle
}

// NewVehicleRepository crea una instancia del repositorio de vehículos en memoria.
func NewVehicleRepository() *VehicleRepository {
	return &VehicleRepository{
		vehicles: make(map[string]*domain.Vehicle),
	}
}

func (r *VehicleRepository) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	if vehicle == nil || vehicle.ID == "" {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.vehicles[vehicle.ID] = vehicle.Clone()
	return nil
}

func (r *VehicleRepository) GetByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, exists := r.vehicles[id]
	if !exists {
		return nil, domain.ErrVehicleNotFound
	}
	return v.Clone(), nil
}

func (r *VehicleRepository) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	if vehicle == nil || vehicle.ID == "" {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.vehicles[vehicle.ID]; !exists {
		return domain.ErrVehicleNotFound
	}

	r.vehicles[vehicle.ID] = vehicle.Clone()
	return nil
}

func (r *VehicleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.vehicles[id]; !exists {
		return domain.ErrVehicleNotFound
	}

	delete(r.vehicles, id)
	return nil
}

func (r *VehicleRepository) List(ctx context.Context, filter domain.VehicleFilter) ([]*domain.Vehicle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Vehicle
	for _, v := range r.vehicles {
		if filter.Brand != "" && !strings.EqualFold(v.Brand, filter.Brand) {
			continue
		}
		if filter.Status != "" && v.Status != filter.Status {
			continue
		}
		if filter.MaxPrice > 0 && v.Price > filter.MaxPrice {
			continue
		}
		if filter.SellerID != "" && v.SellerID != filter.SellerID {
			continue
		}
		result = append(result, v.Clone())
	}

	// Ordenar por fecha de creación descendente (más recientes primero)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}
