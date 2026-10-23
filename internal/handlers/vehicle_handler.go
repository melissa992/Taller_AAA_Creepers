package handlers

import (
	"html"
	"net/http"
	"strings"

	"automarket/internal/domain"
	"automarket/internal/middleware"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

// VehicleHandler gestiona las operaciones de los vehículos del vendedor.
type VehicleHandler struct {
	repo domain.VehicleRepository
}

// NewVehicleHandler inicializa el controlador con la interfaz inyectada.
func NewVehicleHandler(repo domain.VehicleRepository) *VehicleHandler {
	return &VehicleHandler{
		repo: repo,
	}
}

// Dashboard muestra la vista templates/dashboard_seller.html con el formulario y la lista de vehículos propios.
func (h *VehicleHandler) Dashboard(c *gin.Context) {
	sellerID := c.GetString("user_id")
	username := c.GetString("username")

	myVehicles, err := h.repo.GetBySellerID(c.Request.Context(), sellerID)
	if err != nil {
		middleware.SetAuditMetadata(c, "VER_DASHBOARD_VENDEDOR", "/dashboard", "ERROR")
		c.String(http.StatusInternalServerError, "Error consultando vehículos del vendedor")
		return
	}

	middleware.SetAuditMetadata(c, "VER_DASHBOARD_VENDEDOR", "/dashboard", "EXITO")
	c.HTML(http.StatusOK, "dashboard_seller.html", gin.H{
		"Title":      "AutoMarket - Panel del Vendedor",
		"Username":   username,
		"Vehicles":   myVehicles,
		"Total":      len(myVehicles),
		"SuccessMsg": c.Query("success"),
		"ErrorMsg":   c.Query("error"),
	})
}

// Create procesa POST '/vehicles' protegido por Auth y Authz (solo 'Seller').
// Extrae SellerID de gin.Context, sanitiza los campos y guarda el vehículo con estado 'Pending'.
func (h *VehicleHandler) Create(c *gin.Context) {
	sellerID := c.GetString("user_id")
	username := c.GetString("username")

	// Sanitización de inputs contra inyección XSS
	title := sanitize(c.PostForm("title"))
	priceFormatted := sanitize(c.PostForm("price_formatted"))
	yearMileage := sanitize(c.PostForm("year_mileage"))
	location := sanitize(c.PostForm("location"))
	whatsAppNumber := sanitize(c.PostForm("whatsapp_number"))
	imageURL := strings.TrimSpace(c.PostForm("image_url"))

	// Validación de campos
	if title == "" || priceFormatted == "" || yearMileage == "" || location == "" || whatsAppNumber == "" {
		middleware.SetAuditMetadata(c, "CREAR_VEHICULO", "/vehicles", "FALLO_VALIDACION")
		myVehicles, _ := h.repo.GetBySellerID(c.Request.Context(), sellerID)
		c.HTML(http.StatusBadRequest, "dashboard_seller.html", gin.H{
			"Title":          "AutoMarket - Panel del Vendedor",
			"Username":       username,
			"Vehicles":       myVehicles,
			"Total":          len(myVehicles),
			"ErrorMessage":   "Todos los campos marcados con (*) son obligatorios.",
			"TitleVal":       title,
			"PriceVal":       priceFormatted,
			"YearMileageVal": yearMileage,
			"LocationVal":    location,
			"WhatsAppVal":    whatsAppNumber,
			"ImageVal":       imageURL,
		})
		return
	}

	if imageURL == "" {
		imageURL = "/assets/images/car1.jpg"
	}

	// Creación con estado estricto 'Pending' (HU-VEN-01)
	newVehicle := &domain.Vehicle{
		ID:             security.GenerateID("veh"),
		ImageURL:       imageURL,
		Title:          title,
		PriceFormatted: priceFormatted,
		YearMileage:    yearMileage,
		Location:       location,
		WhatsAppNumber: whatsAppNumber,
		Status:         domain.StatusPending,
		SellerID:       sellerID,
	}

	if err := h.repo.Create(c.Request.Context(), newVehicle); err != nil {
		middleware.SetAuditMetadata(c, "CREAR_VEHICULO", newVehicle.ID, "ERROR")
		myVehicles, _ := h.repo.GetBySellerID(c.Request.Context(), sellerID)
		c.HTML(http.StatusInternalServerError, "dashboard_seller.html", gin.H{
			"Title":        "AutoMarket - Panel del Vendedor",
			"Username":     username,
			"Vehicles":     myVehicles,
			"Total":        len(myVehicles),
			"ErrorMessage": "Error interno al guardar el vehículo.",
		})
		return
	}

	middleware.SetAuditMetadata(c, "CREAR_VEHICULO", newVehicle.ID, "EXITO")
	c.Redirect(http.StatusSeeOther, "/dashboard?success=Vehículo+creado+exitosamente.+Se+encuentra+en+estado+Pending+para+revisión.")
}

// sanitize limpia cadenas de texto para evitar inyecciones HTML.
func sanitize(input string) string {
	return html.EscapeString(strings.TrimSpace(input))
}
