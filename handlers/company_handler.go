package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type CompanyHandler struct {
	companyService services.ICompanyService
}

func NewCompanyHandler(companyService services.ICompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
	}
}

func (h *CompanyHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies", h.CreateCompany).Methods(http.MethodPost)
	r.HandleFunc("/companies", h.GetCompanyList).Methods(http.MethodGet)
	r.HandleFunc("/companies/{id}", h.GetCompanyByID).Methods(http.MethodGet)
	r.HandleFunc("/companies/{id}", h.UpdateCompany).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/companies/{id}", h.DeleteCompany).Methods(http.MethodDelete)
}

// CreateCompany godoc
// @Summary Create company
// @Description Create a new company.
// @Tags companies
// @Accept json
// @Produce json
// @Param request body dto.CreateCompanyRequest true "Create company payload"
// @Success 201 {object} utils.APIResponse{data=dto.CompanyResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies [post]
func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCompanyRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.companyService.CreateCompany(r.Context(), &req)
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

// GetCompanyByID godoc
// @Summary Get company by ID
// @Description Retrieve a company by ID.
// @Tags companies
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} utils.APIResponse{data=dto.CompanyResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{id} [get]
func (h *CompanyHandler) GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id")
		return
	}

	result, err := h.companyService.GetCompanyByID(r.Context(), companyID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetCompanyList godoc
// @Summary List companies
// @Description Get paginated list of companies.
// @Tags companies
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Company status"
// @Param search query string false "Search by name or slug"
// @Success 200 {object} utils.APIResponse{data=PaginatedCompanyResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies [get]
func (h *CompanyHandler) GetCompanyList(w http.ResponseWriter, r *http.Request) {
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

	req := &dto.CompanyListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
	}

	result, err := h.companyService.GetCompanyList(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateCompany godoc
// @Summary Update company
// @Description Update a company by ID.
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param request body dto.UpdateCompanyRequest true "Update company payload"
// @Success 200 {object} utils.APIResponse{data=dto.CompanyResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{id} [put]
// @Router /companies/{id} [patch]
func (h *CompanyHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id")
		return
	}

	var req dto.UpdateCompanyRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.companyService.UpdateCompany(r.Context(), companyID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// DeleteCompany godoc
// @Summary Delete company
// @Description Delete a company by ID (soft delete unless hard_delete=true).
// @Tags companies
// @Produce json
// @Param id path string true "Company ID"
// @Param hard_delete query bool false "Hard delete"
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{id} [delete]
func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id")
		return
	}

	hardDelete := r.URL.Query().Get("hard_delete") == "true"
	softDelete := !hardDelete

	if err := h.companyService.DeleteCompany(r.Context(), companyID, softDelete); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "company deleted",
	})
}

func parsePositiveInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, err
	}

	return parsed, nil
}
