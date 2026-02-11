package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type MemoHandler struct {
	memoService services.IMemoService
}

func NewMemoHandler(memoService services.IMemoService) *MemoHandler {
	return &MemoHandler{memoService: memoService}
}

func (h *MemoHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/memos", h.CreateMemo).Methods(http.MethodPost)
	r.HandleFunc("/memos", h.GetMemos).Methods(http.MethodGet)
	r.HandleFunc("/memos/{id}", h.GetMemoByID).Methods(http.MethodGet)
	r.HandleFunc("/memos/{id}/read", h.MarkMemoRead).Methods(http.MethodPost)
	r.HandleFunc("/memos/{id}/approve", h.ApproveMemo).Methods(http.MethodPost)
	r.HandleFunc("/memos/{id}/reject", h.RejectMemo).Methods(http.MethodPost)
}

// CreateMemo godoc
// @Summary Create memo
// @Description Create a new memo and optional recipients.
// @Tags memos
// @Accept json
// @Produce json
// @Param request body dto.CreateMemoRequest true "Create memo payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=models.Memo}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos [post]
func (h *MemoHandler) CreateMemo(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager, roleHRManager, roleSuperAdmin)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id in token")
		return
	}

	var req dto.CreateMemoRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.CompanyID = companyID

	result, err := h.memoService.CreateMemo(r.Context(), employeeID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetMemos godoc
// @Summary List memos
// @Description Get paginated list of memos.
// @Tags memos
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param employeeId query string false "Filter by employee ID"
// @Param status query string false "Filter by status"
// @Param memoType query string false "Filter by memo type"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos [get]
func (h *MemoHandler) GetMemos(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager, roleHRManager, roleSuperAdmin)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company id in token")
		return
	}

	page, err := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, err := parsePositiveInt(r.URL.Query().Get("pageSize"), 10)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid pageSize")
		return
	}

	req := &dto.MemoListRequest{
		Page:      page,
		PageSize:  pageSize,
		CompanyID: companyID,
		Status:    r.URL.Query().Get("status"),
		MemoType:  r.URL.Query().Get("memoType"),
	}

	if employeeIDStr := r.URL.Query().Get("employeeId"); employeeIDStr != "" {
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid employeeId")
			return
		}
		req.EmployeeID = employeeID
	}

	result, err := h.memoService.GetMemos(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetMemoByID godoc
// @Summary Get memo by ID
// @Description Retrieve a memo by ID.
// @Tags memos
// @Produce json
// @Param id path string true "Memo ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.Memo}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos/{id} [get]
func (h *MemoHandler) GetMemoByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleEmployee, roleManager, roleHRManager, roleSuperAdmin); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	memoID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid memo id")
		return
	}

	result, err := h.memoService.GetMemoByID(r.Context(), memoID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// MarkMemoRead godoc
// @Summary Mark memo as read
// @Description Mark a memo as read for the authenticated recipient.
// @Tags memos
// @Produce json
// @Param id path string true "Memo ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos/{id}/read [post]
func (h *MemoHandler) MarkMemoRead(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager, roleHRManager, roleSuperAdmin)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	memoID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid memo id")
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	if err := h.memoService.MarkMemoRead(r.Context(), memoID, employeeID); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "memo marked as read",
	})
}

// ApproveMemo godoc
// @Summary Approve memo
// @Description Approve memo at current workflow step.
// @Tags memos
// @Accept json
// @Produce json
// @Param id path string true "Memo ID"
// @Param request body dto.MemoActionRequest false "Approval comments"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.Memo}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos/{id}/approve [post]
func (h *MemoHandler) ApproveMemo(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleManager, roleHRManager, roleSuperAdmin)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	memoID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid memo id")
		return
	}

	approverID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	var req dto.MemoActionRequest
	_ = utils.DecodeJSONBody(r, &req)

	result, err := h.memoService.ApproveMemo(r.Context(), memoID, approverID, req.Comments)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// RejectMemo godoc
// @Summary Reject memo
// @Description Reject memo at current workflow step.
// @Tags memos
// @Accept json
// @Produce json
// @Param id path string true "Memo ID"
// @Param request body dto.MemoActionRequest false "Rejection comments"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.Memo}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /memos/{id}/reject [post]
func (h *MemoHandler) RejectMemo(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleManager, roleHRManager, roleSuperAdmin)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	memoID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid memo id")
		return
	}

	approverID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	var req dto.MemoActionRequest
	_ = utils.DecodeJSONBody(r, &req)

	result, err := h.memoService.RejectMemo(r.Context(), memoID, approverID, req.Comments)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}
