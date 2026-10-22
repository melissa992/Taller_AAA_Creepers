package handlers

import (
	"net/http"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

// CatalogHandler maneja la vista pública del catálogo de vehículos (HU-VIS-01).
type CatalogHandler struct {
	repo domain.VehicleRepository
}

// NewCatalogHandler construye el handler inyectando el repositorio de vehículos.
func NewCatalogHandler(repo domain.VehicleRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: repo,
	}
}

// Index maneja GET '/' y entrega los vehículos en estado "Published" a la plantilla index.html.
func (h *CatalogHandler) Index(c *gin.Context) {
	vehicles, err := h.repo.GetPublished(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error consultando el catálogo de vehículos")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":    "AutoMarket - Catálogo Oficial de Vehículos",
		"Vehicles": vehicles,
		"Total":    len(vehicles),
	})
}
