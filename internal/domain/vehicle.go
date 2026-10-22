package domain

// VehicleStatus representa el estado del vehículo en AutoMarket.
type VehicleStatus string

const (
	StatusPending   VehicleStatus = "Pending"
	StatusPublished VehicleStatus = "Published"
	StatusSold      VehicleStatus = "Sold"
)

// Vehicle representa la entidad exacta según HU-VIS-01.
type Vehicle struct {
	ID              string        `json:"id"`
	ImageURL        string        `json:"image_url"`
	Title           string        `json:"title"`
	PriceFormatted  string        `json:"price_formatted"`
	YearMileage     string        `json:"year_mileage"`
	Location        string        `json:"location"`
	WhatsAppNumber  string        `json:"whatsapp_number"`
	Status          VehicleStatus `json:"status"`
	SellerID        string        `json:"seller_id"`
}

// Clone produce una copia profunda para seguridad concurrente.
func (v *Vehicle) Clone() *Vehicle {
	if v == nil {
		return nil
	}
	copyVehicle := *v
	return &copyVehicle
}
