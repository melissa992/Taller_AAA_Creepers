package repository_test

import (
	"context"
	"sync"
	"testing"

	"automarket/internal/domain"
	"automarket/internal/repository"
)

func TestMemoryRepositorySingletonAndThreadSafety(t *testing.T) {
	repo1 := repository.GetInstance()
	repo2 := repository.GetInstance()

	if repo1 != repo2 {
		t.Fatal("GetInstance() debe retornar la misma instancia Singleton")
	}

	ctx := context.Background()

	// Verificar datos semilla
	published, err := repo1.GetPublished(ctx)
	if err != nil {
		t.Fatalf("Error consultando vehículos publicados: %v", err)
	}

	if len(published) == 0 {
		t.Fatal("Se esperaban vehículos semilla en estado Published")
	}

	// Verificar que todos los devueltos estén en estado Published y apunten a /assets/images/car1.jpg
	for _, v := range published {
		if v.Status != domain.StatusPublished {
			t.Errorf("El vehículo %s no está en estado Published: %s", v.ID, v.Status)
		}
		if v.ImageURL != "/assets/images/car1.jpg" {
			t.Errorf("El vehículo %s no apunta a /assets/images/car1.jpg: %s", v.ID, v.ImageURL)
		}
		if v.WhatsAppNumber == "" {
			t.Errorf("El vehículo %s no tiene número de WhatsApp", v.ID)
		}
	}

	// Prueba de concurrencia y Thread-Safety con sync.RWMutex
	var wg sync.WaitGroup
	workers := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo1.GetPublished(ctx)
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			veh := &domain.Vehicle{
				ID:             "veh-concurrent",
				ImageURL:       "/assets/images/car1.jpg",
				Title:          "Vehículo Concurrente",
				PriceFormatted: "$10,000 USD",
				YearMileage:    "2022 • 10,000 km",
				Location:       "Cali, Colombia",
				WhatsAppNumber: "573000000000",
				Status:         domain.StatusPublished,
				SellerID:       "usr-seller-1",
			}
			_ = repo1.Create(ctx, veh)
		}(i)
	}

	wg.Wait()
}
