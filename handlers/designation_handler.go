package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type DesignationHandler struct {
	designationService services.IDesignationService
}

func NewDesignationHandler(designationService services.IDesignationService) *DesignationHandler {
	return &DesignationHandler{designationService: designationService}
}

func (h *DesignationHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/designations", h.CreateDesignation).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/designations", h.GetDesignationList).Methods(http.MethodGet)
	r.HandleFunc("/designations/{id}", h.GetDesignationByID).Methods(http.MethodGet)
	r.HandleFunc("/designations/{id}", h.UpdateDesignation).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/designations/{id}", h.DeleteDesignation).Methods(http.MethodDelete)
}

// CreateDesignation godoc
// @Summary Create designation
// @Description Create a new designation.
// @Tags designations
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateDesignationRequest true "Create designation payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=dto.DesignationResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/designations [post]
func (h *DesignationHandler) CreateDesignation(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseEmployeeCompanyID(r, claims)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company_id")
		return
	}

	var req dto.CreateDesignationRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.designationService.CreateDesignation(r.Context(), companyID, &req)
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

// GetDesignationByID godoc
// @Summary Get designation by ID
// @Description Retrieve a designation by ID.
// @Tags designations
// @Produce json
// @Param id path string true "Designation ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.DesignationResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /designations/{id} [get]
func (h *DesignationHandler) GetDesignationByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	designationID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid designation id")
		return
	}

	result, err := h.designationService.GetDesignationByID(r.Context(), designationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetDesignationList godoc
// @Summary List designations
// @Description Get paginated list of designations.
// @Tags designations
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Designation status"
// @Param department_id query string false "Department ID"
// @Param level_id query string false "Level ID"
// @Param search query string false "Search by title or code"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=PaginatedDesignationResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/designations [get]
func (h *DesignationHandler) GetDesignationList(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseEmployeeCompanyID(r, claims)
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

	departmentID := r.URL.Query().Get("department_id")
	if departmentID != "" {
		if _, err := uuid.Parse(departmentID); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid department_id")
			return
		}
	}

	levelID := r.URL.Query().Get("level_id")
	if levelID != "" {
		if _, err := uuid.Parse(levelID); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid level_id")
			return
		}
	}

	req := &dto.DesignationListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		Status:       r.URL.Query().Get("status"),
		DepartmentID: departmentID,
		LevelID:      levelID,
		Search:       r.URL.Query().Get("search"),
	}

	result, err := h.designationService.GetDesignationList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateDesignation godoc
// @Summary Update designation
// @Description Update a designation by ID.
// @Tags designations
// @Accept json
// @Produce json
// @Param id path string true "Designation ID"
// @Param request body dto.UpdateDesignationRequest true "Update designation payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.DesignationResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /designations/{id} [put]
// @Router /designations/{id} [patch]
func (h *DesignationHandler) UpdateDesignation(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	designationID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid designation id")
		return
	}

	var req dto.UpdateDesignationRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.designationService.UpdateDesignation(r.Context(), designationID, &req)
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

// DeleteDesignation godoc
// @Summary Delete designation
// @Description Delete a designation by ID.
// @Tags designations
// @Produce json
// @Param id path string true "Designation ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /designations/{id} [delete]
func (h *DesignationHandler) DeleteDesignation(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	designationID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid designation id")
		return
	}

	if err := h.designationService.DeleteDesignation(r.Context(), designationID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "designation deleted",
	})
}
