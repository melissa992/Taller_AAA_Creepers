package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

// VehicleHandler gestiona las operaciones del catálogo de vehículos en AutoMarket.
type VehicleHandler struct {
	vehicleRepo domain.VehicleRepository
}

// NewVehicleHandler inicializa el controlador inyectando la interfaz del repositorio.
func NewVehicleHandler(vehicleRepo domain.VehicleRepository) *VehicleHandler {
	return &VehicleHandler{
		vehicleRepo: vehicleRepo,
	}
}

// List muestra el catálogo público de vehículos con filtros de búsqueda.
func (h *VehicleHandler) List(c *gin.Context) {
	brandFilter := strings.TrimSpace(c.Query("brand"))
	statusFilter := strings.TrimSpace(c.Query("status"))
	maxPriceStr := strings.TrimSpace(c.Query("max_price"))
	sellerIDFilter := strings.TrimSpace(c.Query("seller_id"))

	var maxPrice float64
	if maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && val > 0 {
			maxPrice = val
		}
	}

	filter := domain.VehicleFilter{
		Brand:    brandFilter,
		Status:   domain.VehicleStatus(statusFilter),
		MaxPrice: maxPrice,
		SellerID: sellerIDFilter,
	}

	vehicles, err := h.vehicleRepo.List(c.Request.Context(), filter)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error Interno",
			"Error": "No se pudo cargar el catálogo de vehículos",
		})
		return
	}

	data := GetCommonContext(c, "Catálogo de Vehículos - AutoMarket")
	data["Vehicles"] = vehicles
	data["FilterBrand"] = brandFilter
	data["FilterStatus"] = statusFilter
	data["FilterMaxPrice"] = maxPriceStr
	data["FilterSellerID"] = sellerIDFilter
	data["Count"] = len(vehicles)

	c.HTML(http.StatusOK, "vehicles/index.html", data)
}

// Show muestra la vista en detalle de un vehículo específico.
func (h *VehicleHandler) Show(c *gin.Context) {
	id := c.Param("id")
	vehicle, err := h.vehicleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVehicleNotFound) {
			c.HTML(http.StatusNotFound, "errors/404.html", gin.H{
				"Title":   "Vehículo No Encontrado",
				"Message": "El vehículo solicitado no existe o fue removido",
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error Interno",
			"Error": "Fallo al consultar el vehículo",
		})
		return
	}

	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("role")
	canManage := (currentUserID == vehicle.SellerID) || (currentUserRole == string(domain.RoleAdmin))

	data := GetCommonContext(c, vehicle.Brand+" "+vehicle.Model+" - AutoMarket")
	data["Vehicle"] = vehicle
	data["CanManage"] = canManage

	c.HTML(http.StatusOK, "vehicles/show.html", data)
}

// ShowCreate muestra el formulario para publicar un nuevo vehículo.
func (h *VehicleHandler) ShowCreate(c *gin.Context) {
	data := GetCommonContext(c, "Publicar Vehículo - AutoMarket")
	c.HTML(http.StatusOK, "vehicles/create.html", data)
}

// Create procesa la creación de un vehículo nuevo publicado por un vendedor o admin.
func (h *VehicleHandler) Create(c *gin.Context) {
	brand := strings.TrimSpace(c.PostForm("brand"))
	model := strings.TrimSpace(c.PostForm("model"))
	yearStr := strings.TrimSpace(c.PostForm("year"))
	priceStr := strings.TrimSpace(c.PostForm("price"))
	mileageStr := strings.TrimSpace(c.PostForm("mileage"))
	description := strings.TrimSpace(c.PostForm("description"))
	imageURL := strings.TrimSpace(c.PostForm("image_url"))

	data := GetCommonContext(c, "Publicar Vehículo - AutoMarket")

	if brand == "" || model == "" || yearStr == "" || priceStr == "" {
		data["ErrorMessage"] = "Marca, modelo, año y precio son obligatorios"
		c.HTML(http.StatusBadRequest, "vehicles/create.html", data)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > time.Now().Year()+1 {
		data["ErrorMessage"] = "El año del vehículo no es válido"
		c.HTML(http.StatusBadRequest, "vehicles/create.html", data)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		data["ErrorMessage"] = "El precio debe ser un número positivo"
		c.HTML(http.StatusBadRequest, "vehicles/create.html", data)
		return
	}

	mileage := 0
	if mileageStr != "" {
		if val, err := strconv.Atoi(mileageStr); err == nil && val >= 0 {
			mileage = val
		}
	}

	if imageURL == "" {
		imageURL = defaultVehicleImage(brand)
	}

	sellerID := c.GetString("user_id")
	sellerName := c.GetString("username")

	newVehicle := &domain.Vehicle{
		ID:          security.GenerateID("veh"),
		Brand:       brand,
		Model:       model,
		Year:        year,
		Price:       price,
		Mileage:     mileage,
		Status:      domain.StatusAvailable,
		Description: description,
		SellerID:    sellerID,
		SellerName:  sellerName,
		ImageURL:    imageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.vehicleRepo.Create(c.Request.Context(), newVehicle); err != nil {
		data["ErrorMessage"] = "Error al guardar el vehículo"
		c.HTML(http.StatusInternalServerError, "vehicles/create.html", data)
		return
	}

	c.Redirect(http.StatusSeeOther, "/vehicles/"+newVehicle.ID+"?success=vehiculo_creado")
}

// ShowEdit presenta la vista de edición verificando propiedad de recurso (Ownership Authorization).
func (h *VehicleHandler) ShowEdit(c *gin.Context) {
	id := c.Param("id")
	vehicle, err := h.vehicleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "errors/404.html", gin.H{
			"Title":   "Vehículo No Encontrado",
			"Message": "El vehículo no existe",
		})
		return
	}

	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("role")

	// Control de autorización a nivel de recurso (Ownership RBAC)
	if vehicle.SellerID != currentUserID && currentUserRole != string(domain.RoleAdmin) {
		c.HTML(http.StatusForbidden, "errors/403.html", gin.H{
			"Title":    "Acceso Denegado (403)",
			"UserRole": currentUserRole,
			"Message":  "No tiene autorización para editar un vehículo publicado por otro usuario.",
		})
		return
	}

	data := GetCommonContext(c, "Editar Vehículo - AutoMarket")
	data["Vehicle"] = vehicle
	c.HTML(http.StatusOK, "vehicles/edit.html", data)
}

// Update procesa la actualización de datos de un vehículo existente.
func (h *VehicleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	vehicle, err := h.vehicleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "errors/404.html", gin.H{
			"Title":   "No Encontrado",
			"Message": "El vehículo a actualizar no existe",
		})
		return
	}

	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("role")

	if vehicle.SellerID != currentUserID && currentUserRole != string(domain.RoleAdmin) {
		c.HTML(http.StatusForbidden, "errors/403.html", gin.H{
			"Title":    "Acceso Denegado (403)",
			"UserRole": currentUserRole,
			"Message":  "No tiene autorización para modificar este recurso.",
		})
		return
	}

	brand := strings.TrimSpace(c.PostForm("brand"))
	model := strings.TrimSpace(c.PostForm("model"))
	yearStr := strings.TrimSpace(c.PostForm("year"))
	priceStr := strings.TrimSpace(c.PostForm("price"))
	mileageStr := strings.TrimSpace(c.PostForm("mileage"))
	statusStr := strings.TrimSpace(c.PostForm("status"))
	description := strings.TrimSpace(c.PostForm("description"))
	imageURL := strings.TrimSpace(c.PostForm("image_url"))

	data := GetCommonContext(c, "Editar Vehículo - AutoMarket")
	data["Vehicle"] = vehicle

	if brand == "" || model == "" || yearStr == "" || priceStr == "" {
		data["ErrorMessage"] = "Marca, modelo, año y precio son obligatorios"
		c.HTML(http.StatusBadRequest, "vehicles/edit.html", data)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > time.Now().Year()+1 {
		data["ErrorMessage"] = "El año del vehículo no es válido"
		c.HTML(http.StatusBadRequest, "vehicles/edit.html", data)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		data["ErrorMessage"] = "El precio debe ser un número mayor a cero"
		c.HTML(http.StatusBadRequest, "vehicles/edit.html", data)
		return
	}

	mileage := vehicle.Mileage
	if mileageStr != "" {
		if val, err := strconv.Atoi(mileageStr); err == nil && val >= 0 {
			mileage = val
		}
	}

	var status domain.VehicleStatus
	switch statusStr {
	case string(domain.StatusReserved):
		status = domain.StatusReserved
	case string(domain.StatusSold):
		status = domain.StatusSold
	default:
		status = domain.StatusAvailable
	}

	if imageURL == "" {
		imageURL = vehicle.ImageURL
	}

	vehicle.Brand = brand
	vehicle.Model = model
	vehicle.Year = year
	vehicle.Price = price
	vehicle.Mileage = mileage
	vehicle.Status = status
	vehicle.Description = description
	vehicle.ImageURL = imageURL
	vehicle.UpdatedAt = time.Now()

	if err := h.vehicleRepo.Update(c.Request.Context(), vehicle); err != nil {
		data["ErrorMessage"] = "Error al actualizar el vehículo"
		c.HTML(http.StatusInternalServerError, "vehicles/edit.html", data)
		return
	}

	c.Redirect(http.StatusSeeOther, "/vehicles/"+vehicle.ID+"?success=vehiculo_actualizado")
}

// Delete elimina un vehículo verificando la autorización de propiedad.
func (h *VehicleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	vehicle, err := h.vehicleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/vehicles?error=vehiculo_no_encontrado")
		return
	}

	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("role")

	if vehicle.SellerID != currentUserID && currentUserRole != string(domain.RoleAdmin) {
		c.HTML(http.StatusForbidden, "errors/403.html", gin.H{
			"Title":    "Acceso Denegado (403)",
			"UserRole": currentUserRole,
			"Message":  "No tiene permisos para eliminar este vehículo.",
		})
		return
	}

	if err := h.vehicleRepo.Delete(c.Request.Context(), id); err != nil {
		c.Redirect(http.StatusSeeOther, "/vehicles/"+id+"?error=error_al_eliminar")
		return
	}

	c.Redirect(http.StatusSeeOther, "/vehicles?success=vehiculo_eliminado")
}

// defaultVehicleImage retorna una imagen temátizada según la marca.
func defaultVehicleImage(brand string) string {
	b := strings.ToLower(brand)
	switch {
	case strings.Contains(b, "toyota"):
		return "https://images.unsplash.com/photo-1621007947382-bb3c3994e3fb?auto=format&fit=crop&w=800&q=80"
	case strings.Contains(b, "bmw"):
		return "https://images.unsplash.com/photo-1555215695-3004980ad54e?auto=format&fit=crop&w=800&q=80"
	case strings.Contains(b, "ford"):
		return "https://images.unsplash.com/photo-1552519507-da3b142c6e3d?auto=format&fit=crop&w=800&q=80"
	case strings.Contains(b, "porsche"):
		return "https://images.unsplash.com/photo-1614162692292-7ac56d7f7f1e?auto=format&fit=crop&w=800&q=80"
	case strings.Contains(b, "audi"):
		return "https://images.unsplash.com/photo-1603584173870-7f23fdae1b7a?auto=format&fit=crop&w=800&q=80"
	default:
		return "https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=800&q=80"
	}
}
