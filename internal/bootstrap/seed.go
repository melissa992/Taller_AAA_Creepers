package bootstrap

import (
	"context"
	"fmt"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"
)

// SeedInitialData puebla los repositorios en memoria con usuarios y vehículos de prueba.
func SeedInitialData(
	userRepo domain.UserRepository,
	vehicleRepo domain.VehicleRepository,
) error {
	ctx := context.Background()

	// 1. Usuarios de Prueba
	usersToSeed := []struct {
		username string
		email    string
		password string
		role     domain.Role
	}{
		{"admin", "admin@automarket.local", "Admin123!", domain.RoleAdmin},
		{"vendedor1", "carlos.vendedor@automarket.local", "Vendedor123!", domain.RoleSeller},
		{"comprador1", "ana.compradora@automarket.local", "Comprador123!", domain.RoleBuyer},
	}

	createdUsers := make(map[string]*domain.User)

	for _, u := range usersToSeed {
		salt, err := security.GenerateSalt()
		if err != nil {
			return fmt.Errorf("error generando sal: %w", err)
		}
		hash := security.HashPassword(u.password, salt)

		user := &domain.User{
			ID:           security.GenerateID("usr"),
			Username:     u.username,
			Email:        u.email,
			PasswordHash: hash,
			Salt:         salt,
			Role:         u.role,
			CreatedAt:    time.Now().Add(-48 * time.Hour),
		}

		if err := userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("error creando usuario inicial %s: %w", u.username, err)
		}
		createdUsers[u.username] = user
	}

	seller := createdUsers["vendedor1"]
	admin := createdUsers["admin"]

	// 2. Catálogo Inicial de Vehículos
	vehiclesToSeed := []*domain.Vehicle{
		{
			ID:          "veh-toyota-lc300",
			Brand:       "Toyota",
			Model:       "Land Cruiser 300 GR-Sport",
			Year:        2024,
			Price:       135000,
			Mileage:     4500,
			Status:      domain.StatusAvailable,
			Description: "Edición GR Sport con suspensión adaptativa E-KDSS, motor V6 Twin Turbo y paquete off-road completo. Único dueño, garantía de fábrica.",
			SellerID:    seller.ID,
			SellerName:  seller.Username,
			ImageURL:    "https://images.unsplash.com/photo-1621007947382-bb3c3994e3fb?auto=format&fit=crop&w=1000&q=80",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "veh-bmw-m3",
			Brand:       "BMW",
			Model:       "M3 Competition xDrive",
			Year:        2023,
			Price:       98500,
			Mileage:     12400,
			Status:      domain.StatusAvailable,
			Description: "Motor S58 3.0L Twin-Turbo 503 HP, tracción integral M xDrive, frenos cerámicos de carbono y asientos M Carbon bucket en piel Merino.",
			SellerID:    seller.ID,
			SellerName:  seller.Username,
			ImageURL:    "https://images.unsplash.com/photo-1555215695-3004980ad54e?auto=format&fit=crop&w=1000&q=80",
			CreatedAt:   time.Now().Add(-18 * time.Hour),
			UpdatedAt:   time.Now().Add(-18 * time.Hour),
		},
		{
			ID:          "veh-porsche-911",
			Brand:       "Porsche",
			Model:       "911 GT3 Touring",
			Year:        2024,
			Price:       225000,
			Mileage:     2300,
			Status:      domain.StatusReserved,
			Description: "Motor bóxer atmosférico 4.0L de 502 HP, transmisión manual de 6 velocidades, paquete Touring sin alerón fijo, color Verde Roble metalizado.",
			SellerID:    admin.ID,
			SellerName:  admin.Username,
			ImageURL:    "https://images.unsplash.com/photo-1614162692292-7ac56d7f7f1e?auto=format&fit=crop&w=1000&q=80",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now().Add(-6 * time.Hour),
		},
		{
			ID:          "veh-ford-mustang",
			Brand:       "Ford",
			Model:       "Mustang Dark Horse 5.0L",
			Year:        2024,
			Price:       68900,
			Mileage:     6100,
			Status:      domain.StatusAvailable,
			Description: "V8 Coyote 500 HP con caja manual Tremec de 6 marchas, amortiguadores MagneRide, escape activo y paquete de apariencia Dark Horse.",
			SellerID:    seller.ID,
			SellerName:  seller.Username,
			ImageURL:    "https://images.unsplash.com/photo-1552519507-da3b142c6e3d?auto=format&fit=crop&w=1000&q=80",
			CreatedAt:   time.Now().Add(-8 * time.Hour),
			UpdatedAt:   time.Now().Add(-8 * time.Hour),
		},
		{
			ID:          "veh-audi-rs6",
			Brand:       "Audi",
			Model:       "RS6 Avant Performance",
			Year:        2023,
			Price:       142000,
			Mileage:     18000,
			Status:      domain.StatusSold,
			Description: "V8 4.0L Bi-Turbo con 621 HP, tracción quattro, diferencial deportivo, faros HD Matrix LED láser y sistema de sonido Bang & Olufsen 3D.",
			SellerID:    admin.ID,
			SellerName:  admin.Username,
			ImageURL:    "https://images.unsplash.com/photo-1603584173870-7f23fdae1b7a?auto=format&fit=crop&w=1000&q=80",
			CreatedAt:   time.Now().Add(-30 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}

	for _, v := range vehiclesToSeed {
		if err := vehicleRepo.Create(ctx, v); err != nil {
			return fmt.Errorf("error creando vehículo inicial %s: %w", v.Model, err)
		}
	}

	return nil
}
