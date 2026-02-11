package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type ApprovalHandler struct {
	approvalService services.IApprovalService
}

func NewApprovalHandler(approvalService services.IApprovalService) *ApprovalHandler {
	return &ApprovalHandler{approvalService: approvalService}
}

func (h *ApprovalHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/approval-workflows", h.CreateApprovalWorkflow).Methods(http.MethodPost)
	r.HandleFunc("/approval-workflows", h.GetApprovalWorkflowList).Methods(http.MethodGet)
	r.HandleFunc("/approval-history", h.GetApprovalHistory).Methods(http.MethodGet)
}

// CreateApprovalWorkflow godoc
// @Summary Create approval workflow
// @Description Create an approval workflow for a company.
// @Tags approval-workflows
// @Accept json
// @Produce json
// @Param request body dto.CreateApprovalWorkflowRequest true "Create approval workflow payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=models.ApprovalWorkflow}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /approval-workflows [post]
func (h *ApprovalHandler) CreateApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req dto.CreateApprovalWorkflowRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CompanyID == uuid.Nil {
		companyID, err := uuid.Parse(claims.CompanyID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid company id")
			return
		}
		req.CompanyID = companyID
	}

	result, err := h.approvalService.CreateWorkflow(r.Context(), &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetApprovalWorkflowList godoc
// @Summary List approval workflows
// @Description Get approval workflows for the authenticated company.
// @Tags approval-workflows
// @Produce json
// @Param workflowType query string false "Workflow type (leave/memo/expense)"
// @Param departmentId query string false "Department ID filter"
// @Param onlyActive query boolean false "Only active workflows (default true)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]models.ApprovalWorkflow}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /approval-workflows [get]
func (h *ApprovalHandler) GetApprovalWorkflowList(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id in token")
		return
	}

	req := &dto.ApprovalWorkflowListRequest{
		CompanyID:    companyID,
		WorkflowType: r.URL.Query().Get("workflowType"),
		OnlyActive:   r.URL.Query().Get("onlyActive") != "false",
	}

	if departmentID := r.URL.Query().Get("departmentId"); departmentID != "" {
		parsedDepartmentID, err := uuid.Parse(departmentID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid departmentId")
			return
		}
		req.DepartmentID = parsedDepartmentID
	}

	result, err := h.approvalService.GetWorkflowList(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetApprovalHistory godoc
// @Summary Get approval history
// @Description Get approval history entries for an entity.
// @Tags approval-history
// @Produce json
// @Param entityType query string true "Entity type (leave_request/memo)"
// @Param entityId query string true "Entity ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]models.ApprovalHistory}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /approval-history [get]
func (h *ApprovalHandler) GetApprovalHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleManager, roleEmployee); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	entityType := r.URL.Query().Get("entityType")
	entityIDStr := r.URL.Query().Get("entityId")
	if entityType == "" || entityIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "entityType and entityId are required")
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid entityId")
		return
	}

	result, err := h.approvalService.GetApprovalHistory(r.Context(), entityType, entityID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}
