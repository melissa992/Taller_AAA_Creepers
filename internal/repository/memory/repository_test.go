package memory_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"automarket/internal/domain"
	"automarket/internal/repository/memory"
)

func TestUserConcurrentOperations(t *testing.T) {
	repo := memory.NewUserRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	workers := 50

	// Escritura concurrente
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user := &domain.User{
				ID:        fmt.Sprintf("usr-%d", id),
				Username:  fmt.Sprintf("user%d", id),
				Email:     fmt.Sprintf("user%d@test.com", id),
				Role:      domain.RoleBuyer,
				CreatedAt: time.Now(),
			}
			_ = repo.Create(ctx, user)
		}(i)
	}

	// Lectura concurrente simultánea
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = repo.GetByUsername(ctx, fmt.Sprintf("user%d", id))
			_, _ = repo.ListAll(ctx)
		}(i)
	}

	wg.Wait()

	users, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("Error consultando usuarios: %v", err)
	}

	if len(users) != workers {
		t.Errorf("Se esperaban %d usuarios, obtenidos: %d", workers, len(users))
	}
}

func TestVehicleConcurrentOperations(t *testing.T) {
	repo := memory.NewVehicleRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	total := 40

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			veh := &domain.Vehicle{
				ID:       fmt.Sprintf("veh-%d", id),
				Brand:    "Toyota",
				Model:    fmt.Sprintf("Model-%d", id),
				Year:     2022,
				Price:    float64(20000 + id*500),
				Status:   domain.StatusAvailable,
				SellerID: "seller-1",
			}
			_ = repo.Create(ctx, veh)
		}(i)
	}

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.List(ctx, domain.VehicleFilter{Brand: "Toyota"})
		}()
	}

	wg.Wait()

	list, err := repo.List(ctx, domain.VehicleFilter{})
	if err != nil {
		t.Fatalf("Error listando vehículos: %v", err)
	}

	if len(list) != total {
		t.Errorf("Se esperaban %d vehículos, obtenidos: %d", total, len(list))
	}
}
