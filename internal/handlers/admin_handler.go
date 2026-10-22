package handlers

import (
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AdminHandler gestiona la moderación de publicaciones y usuarios (HU-ADM-01).
type AdminHandler struct {
	vehicleRepo domain.VehicleRepository
	userRepo    domain.UserRepository
	auditRepo   domain.AuditRepository
}

// NewAdminHandler inyecta las dependencias necesarias para el administrador.
func NewAdminHandler(
	vehicleRepo domain.VehicleRepository,
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
) *AdminHandler {
	return &AdminHandler{
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
	}
}

// ShowPublications renderiza templates/admin_publicaciones.html con las publicaciones pendientes y el listado general.
func (h *AdminHandler) ShowPublications(c *gin.Context) {
	ctx := c.Request.Context()
	username := c.GetString("username")

	allVehicles, err := h.vehicleRepo.ListAll(ctx)
	if err != nil {
		middleware.SetAuditMetadata(c, "CONSULTAR_PUBLICACIONES_ADMIN", "vehicles", "ERROR")
		c.String(http.StatusInternalServerError, "Error consultando las publicaciones")
		return
	}

	var pendingVehicles []*domain.Vehicle
	for _, v := range allVehicles {
		if v.Status == domain.StatusPending {
			pendingVehicles = append(pendingVehicles, v)
		}
	}

	middleware.SetAuditMetadata(c, "CONSULTAR_PUBLICACIONES_ADMIN", "vehicles", "EXITO")

	c.HTML(http.StatusOK, "admin_publicaciones.html", gin.H{
		"Title":           "AutoMarket - Moderación de Publicaciones",
		"Username":        username,
		"PendingVehicles": pendingVehicles,
		"AllVehicles":     allVehicles,
		"PendingCount":    len(pendingVehicles),
		"TotalCount":      len(allVehicles),
		"SuccessMsg":      c.Query("success"),
		"ErrorMsg":        c.Query("error"),
	})
}

// ApprovePublication cambia una publicación de estado Pending a Published (HU-ADM-01).
func (h *AdminHandler) ApprovePublication(c *gin.Context) {
	ctx := c.Request.Context()
	vehicleID := c.Param("id")

	// Control de integridad antes de ejecutar operación crítica (Tarea 5)
	if ok, _, err := h.auditRepo.VerifyIntegrity(ctx); !ok || err != nil {
		middleware.SetAuditMetadata(c, "APROBAR_PUBLICACION", vehicleID, "ERROR_INTEGRIDAD_AUDITORIA")
		c.String(http.StatusInternalServerError, "Operación bloqueada: fallo de integridad en los registros de auditoría")
		return
	}

	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		middleware.SetAuditMetadata(c, "APROBAR_PUBLICACION", vehicleID, "NO_ENCONTRADO")
		c.Redirect(http.StatusSeeOther, "/admin/publicaciones?error=La+publicación+no+existe")
		return
	}

	if vehicle.Status != domain.StatusPending {
		middleware.SetAuditMetadata(c, "APROBAR_PUBLICACION", vehicleID, "ESTADO_INVALIDO")
		c.Redirect(http.StatusSeeOther, "/admin/publicaciones?error=Solo+se+pueden+aprobar+publicaciones+pendientes")
		return
	}

	vehicle.Status = domain.StatusPublished
	if err := h.vehicleRepo.Update(ctx, vehicle); err != nil {
		middleware.SetAuditMetadata(c, "APROBAR_PUBLICACION", vehicleID, "ERROR")
		c.Redirect(http.StatusSeeOther, "/admin/publicaciones?error=Error+al+aprobar+la+publicación")
		return
	}

	middleware.SetAuditMetadata(c, "APROBAR_PUBLICACION", vehicleID, "EXITO")
	c.Redirect(http.StatusSeeOther, "/admin/publicaciones?success=Publicación+aprobada+y+visible+en+el+catálogo+público")
}

// DeletePublication elimina una publicación del sistema (HU-ADM-01).
func (h *AdminHandler) DeletePublication(c *gin.Context) {
	ctx := c.Request.Context()
	vehicleID := c.Param("id")

	if ok, _, err := h.auditRepo.VerifyIntegrity(ctx); !ok || err != nil {
		middleware.SetAuditMetadata(c, "ELIMINAR_PUBLICACION", vehicleID, "ERROR_INTEGRIDAD_AUDITORIA")
		c.String(http.StatusInternalServerError, "Operación bloqueada: fallo de integridad en los registros de auditoría")
		return
	}

	if err := h.vehicleRepo.Delete(ctx, vehicleID); err != nil {
		middleware.SetAuditMetadata(c, "ELIMINAR_PUBLICACION", vehicleID, "NO_ENCONTRADO")
		c.Redirect(http.StatusSeeOther, "/admin/publicaciones?error=No+se+encontró+la+publicación+a+eliminar")
		return
	}

	middleware.SetAuditMetadata(c, "ELIMINAR_PUBLICACION", vehicleID, "EXITO")
	c.Redirect(http.StatusSeeOther, "/admin/publicaciones?success=Publicación+eliminada+correctamente")
}

// ShowUsers renderiza templates/admin_usuarios.html con el listado de usuarios.
func (h *AdminHandler) ShowUsers(c *gin.Context) {
	ctx := c.Request.Context()
	username := c.GetString("username")
	currentUserID := c.GetString("user_id")

	users, err := h.userRepo.ListAllUsers(ctx)
	if err != nil {
		middleware.SetAuditMetadata(c, "CONSULTAR_USUARIOS_ADMIN", "users", "ERROR")
		c.String(http.StatusInternalServerError, "Error consultando usuarios")
		return
	}

	middleware.SetAuditMetadata(c, "CONSULTAR_USUARIOS_ADMIN", "users", "EXITO")

	c.HTML(http.StatusOK, "admin_usuarios.html", gin.H{
		"Title":         "AutoMarket - Administración de Usuarios",
		"Username":      username,
		"CurrentUserID": currentUserID,
		"Users":         users,
		"Total":         len(users),
		"SuccessMsg":    c.Query("success"),
		"ErrorMsg":      c.Query("error"),
	})
}

// DeleteUser elimina un usuario, sus publicaciones asociadas y sus sesiones (HU-ADM-01).
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	targetUserID := c.Param("id")
	currentAdminID := c.GetString("user_id")

	if ok, _, err := h.auditRepo.VerifyIntegrity(ctx); !ok || err != nil {
		middleware.SetAuditMetadata(c, "ELIMINAR_USUARIO", targetUserID, "ERROR_INTEGRIDAD_AUDITORIA")
		c.String(http.StatusInternalServerError, "Operación bloqueada: fallo de integridad en los registros de auditoría")
		return
	}

	// Evitar que el administrador en sesión se elimine a sí mismo
	if targetUserID == currentAdminID {
		middleware.SetAuditMetadata(c, "ELIMINAR_USUARIO", targetUserID, "OPERACION_NO_PERMITIDA")
		c.Redirect(http.StatusSeeOther, "/admin/usuarios?error=No+puedes+eliminar+tu+propia+cuenta+de+administrador")
		return
	}

	if err := h.userRepo.DeleteUser(ctx, targetUserID); err != nil {
		middleware.SetAuditMetadata(c, "ELIMINAR_USUARIO", targetUserID, "NO_ENCONTRADO")
		c.Redirect(http.StatusSeeOther, "/admin/usuarios?error=El+usuario+no+existe")
		return
	}

	middleware.SetAuditMetadata(c, "ELIMINAR_USUARIO", targetUserID, "EXITO")
	c.Redirect(http.StatusSeeOther, "/admin/usuarios?success=Usuario+y+sus+publicaciones+asociadas+eliminados+correctamente")
}
