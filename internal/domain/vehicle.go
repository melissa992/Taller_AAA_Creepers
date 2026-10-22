package domain

import "time"

// VehicleStatus representa el estado comercial de un vehículo.
type VehicleStatus string

const (
	StatusAvailable VehicleStatus = "AVAILABLE"
	StatusReserved  VehicleStatus = "RESERVED"
	StatusSold      VehicleStatus = "SOLD"
)

// Vehicle representa la entidad pura de un vehículo en AutoMarket.
type Vehicle struct {
	ID          string        `json:"id"`
	Brand       string        `json:"brand"`
	Model       string        `json:"model"`
	Year        int           `json:"year"`
	Price       float64       `json:"price"`
	Mileage     int           `json:"mileage"`
	Status      VehicleStatus `json:"status"`
	Description string        `json:"description"`
	SellerID    string        `json:"seller_id"`
	SellerName  string        `json:"seller_name"`
	ImageURL    string        `json:"image_url"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Clone produce una copia segura para concurrencia en memoria.
func (v *Vehicle) Clone() *Vehicle {
	if v == nil {
		return nil
	}
	copyVehicle := *v
	return &copyVehicle
}

// VehicleFilter define los criterios de búsqueda en el repositorio.
type VehicleFilter struct {
	Brand    string
	Status   VehicleStatus
	MaxPrice float64
	SellerID string
}
