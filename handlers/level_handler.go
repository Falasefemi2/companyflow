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

type LevelHandler struct {
	levelService services.ILevelService
}

func NewLevelHandler(levelService services.ILevelService) *LevelHandler {
	return &LevelHandler{levelService: levelService}
}

func (h *LevelHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/levels", h.CreateLevel).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/levels", h.GetLevelList).Methods(http.MethodGet)
	r.HandleFunc("/levels/{id}", h.GetLevelByID).Methods(http.MethodGet)
	r.HandleFunc("/levels/{id}", h.UpdateLevel).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/levels/{id}", h.DeleteLevel).Methods(http.MethodDelete)
}

// CreateLevel godoc
// @Summary Create level
// @Description Create a new level.
// @Tags levels
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateLevelRequest true "Create level payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=dto.LevelResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/levels [post]
func (h *LevelHandler) CreateLevel(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateLevelRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.levelService.CreateLevel(r.Context(), companyID, &req)
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

// GetLevelByID godoc
// @Summary Get level by ID
// @Description Retrieve a level by ID.
// @Tags levels
// @Produce json
// @Param id path string true "Level ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.LevelResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /levels/{id} [get]
func (h *LevelHandler) GetLevelByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	levelID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid level id")
		return
	}

	result, err := h.levelService.GetLevelByID(r.Context(), levelID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLevelList godoc
// @Summary List levels
// @Description Get paginated list of levels.
// @Tags levels
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by name or code"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=PaginatedLevelResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/levels [get]
func (h *LevelHandler) GetLevelList(w http.ResponseWriter, r *http.Request) {
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

	req := &dto.LevelListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		Search: r.URL.Query().Get("search"),
	}

	result, err := h.levelService.GetLevelList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateLevel godoc
// @Summary Update level
// @Description Update a level by ID.
// @Tags levels
// @Accept json
// @Produce json
// @Param id path string true "Level ID"
// @Param request body dto.UpdateLevelRequest true "Update level payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.LevelResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /levels/{id} [put]
// @Router /levels/{id} [patch]
func (h *LevelHandler) UpdateLevel(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	levelID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid level id")
		return
	}

	var req dto.UpdateLevelRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.levelService.UpdateLevel(r.Context(), levelID, &req)
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

// DeleteLevel godoc
// @Summary Delete level
// @Description Delete a level by ID.
// @Tags levels
// @Produce json
// @Param id path string true "Level ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /levels/{id} [delete]
func (h *LevelHandler) DeleteLevel(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	levelID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid level id")
		return
	}

	if err := h.levelService.DeleteLevel(r.Context(), levelID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "level deleted",
	})
}
