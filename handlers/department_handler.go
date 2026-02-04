package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

const (
	roleSuperAdmin = "Super Admin"
	roleHRManager  = "HR Manager"
)

type DepartmentHandler struct {
	departmentService services.IDepartmentService
}

func NewDepartmentHandler(departmentService services.IDepartmentService) *DepartmentHandler {
	return &DepartmentHandler{departmentService: departmentService}
}

func (h *DepartmentHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/departments", h.CreateDepartment).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/departments", h.GetDepartmentList).Methods(http.MethodGet)
	r.HandleFunc("/departments/{id}", h.GetDepartmentByID).Methods(http.MethodGet)
	r.HandleFunc("/departments/{id}", h.UpdateDepartment).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/departments/{id}", h.DeleteDepartment).Methods(http.MethodDelete)
}

// CreateDepartment godoc
// @Summary Create department
// @Description Create a new department.
// @Tags departments
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateDepartmentRequest true "Create department payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=dto.DepartmentResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/departments [post]
func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseCompanyID(r, claims)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company_id")
		return
	}

	var req dto.CreateDepartmentRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.departmentService.CreateDepartment(r.Context(), companyID, &req)
	if err != nil {
		var vErr *utils.ValidationError
		if errors.As(err, &vErr) {
			utils.RespondWithError(w, http.StatusBadRequest, vErr.Message)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetDepartmentByID godoc
// @Summary Get department by ID
// @Description Retrieve a department by ID.
// @Tags departments
// @Produce json
// @Param id path string true "Department ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.DepartmentResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /departments/{id} [get]
func (h *DepartmentHandler) GetDepartmentByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	departmentID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid department id")
		return
	}

	result, err := h.departmentService.GetDepartmentByID(r.Context(), departmentID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetDepartmentList godoc
// @Summary List departments
// @Description Get paginated list of departments.
// @Tags departments
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Department status"
// @Param search query string false "Search by name or code"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=PaginatedDepartmentResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/departments [get]
func (h *DepartmentHandler) GetDepartmentList(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseCompanyID(r, claims)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company_id")
		return
	}

	page, err := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid page")
		return
	}

	pageSize, err := parsePositiveInt(r.URL.Query().Get("page_size"), 10)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid page_size")
		return
	}

	req := &dto.DepartmentListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
	}

	result, err := h.departmentService.GetDepartmentList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateDepartment godoc
// @Summary Update department
// @Description Update a department by ID.
// @Tags departments
// @Accept json
// @Produce json
// @Param id path string true "Department ID"
// @Param request body dto.UpdateDepartmentRequest true "Update department payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.DepartmentResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /departments/{id} [put]
// @Router /departments/{id} [patch]
func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	departmentID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid department id")
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.departmentService.UpdateDepartment(r.Context(), departmentID, &req)
	if err != nil {
		var vErr *utils.ValidationError
		if errors.As(err, &vErr) {
			utils.RespondWithError(w, http.StatusBadRequest, vErr.Message)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// DeleteDepartment godoc
// @Summary Delete department
// @Description Delete a department by ID (soft delete unless hard_delete=true).
// @Tags departments
// @Produce json
// @Param id path string true "Department ID"
// @Param hard_delete query bool false "Hard delete"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /departments/{id} [delete]
func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	departmentID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid department id")
		return
	}

	hardDelete := r.URL.Query().Get("hard_delete") == "true"
	softDelete := !hardDelete

	if err := h.departmentService.DeleteDepartment(r.Context(), departmentID, softDelete); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "department deleted",
	})
}

func parseCompanyID(r *http.Request, claims *utils.AuthClaims) (uuid.UUID, error) {
	companyIDStr := mux.Vars(r)["company_id"]
	if companyIDStr == "" {
		companyIDStr = claims.CompanyID
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	if claims.CompanyID != "" && claims.CompanyID != companyID.String() {
		return uuid.Nil, errors.New("company_id mismatch")
	}

	return companyID, nil
}

func authorizeToken(r *http.Request, allowedRoles ...string) (*utils.AuthClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid authorization header")
	}

	claims, err := utils.ValidateToken(parts[1])
	if err != nil {
		return nil, errors.New("invalid token")
	}

	for _, allowed := range allowedRoles {
		if claims.Role == allowed {
			return claims, nil
		}
	}

	return nil, errors.New("insufficient role")
}

// func parsePositiveInt(value string, defaultValue int) (int, error) {
// 	if value == "" {
// 		return defaultValue, nil
// 	}

// 	parsed, err := strconv.Atoi(value)
// 	if err != nil || parsed <= 0 {
// 		return 0, err
// 	}

// 	return parsed, nil
// }
