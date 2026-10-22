package handlers

import (
	"net/http"

	"automarket/internal/domain"

	"github.com/gin-gonic/gin"
)

// DashboardHandler coordina la vista de panel de control y visor de auditoría AAA.
type DashboardHandler struct {
	vehicleRepo domain.VehicleRepository
	auditRepo   domain.AuditRepository
	userRepo    domain.UserRepository
}

// NewDashboardHandler inicializa el controlador con dependencias inyectadas.
func NewDashboardHandler(
	vehicleRepo domain.VehicleRepository,
	auditRepo domain.AuditRepository,
	userRepo domain.UserRepository,
) *DashboardHandler {
	return &DashboardHandler{
		vehicleRepo: vehicleRepo,
		auditRepo:   auditRepo,
		userRepo:    userRepo,
	}
}

// Index muestra el panel de control según el rol del usuario autenticado.
func (h *DashboardHandler) Index(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("role")

	data := GetCommonContext(c, "Panel de Control - AutoMarket")

	allVehicles, _ := h.vehicleRepo.List(c.Request.Context(), domain.VehicleFilter{})
	data["TotalVehicles"] = len(allVehicles)

	myVehicles, _ := h.vehicleRepo.List(c.Request.Context(), domain.VehicleFilter{SellerID: currentUserID})
	data["MyVehicles"] = myVehicles
	data["MyVehiclesCount"] = len(myVehicles)

	if currentUserRole == string(domain.RoleAdmin) {
		users, _ := h.userRepo.ListAll(c.Request.Context())
		data["UsersList"] = users
		data["TotalUsers"] = len(users)

		recentLogs, _ := h.auditRepo.ListRecent(c.Request.Context(), 10)
		data["RecentLogs"] = recentLogs
	}

	c.HTML(http.StatusOK, "dashboard/index.html", data)
}

// AuditLogs presenta el visor de trazabilidad y auditoría completa (Accounting en AAA).
// Exclusivo para administradores.
func (h *DashboardHandler) AuditLogs(c *gin.Context) {
	logs, err := h.auditRepo.ListRecent(c.Request.Context(), 100)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
			"Title": "Error Interno",
			"Error": "No se pudieron obtener los registros de auditoría",
		})
		return
	}

	data := GetCommonContext(c, "Auditoría AAA (Accounting) - AutoMarket")
	data["Logs"] = logs
	data["TotalLogs"] = len(logs)

	c.HTML(http.StatusOK, "dashboard/audit.html", data)
}
