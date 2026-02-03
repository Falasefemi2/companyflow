package handlers

import (
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

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCompanyRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.companyService.CreateCompany(r.Context(), &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

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
