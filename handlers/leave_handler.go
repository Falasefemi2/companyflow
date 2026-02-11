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

const (
	roleEmployee = "Employee"
	roleManager  = "Manager"
)

type LeaveHandler struct {
	leaveService services.ILeaveService
}

func NewLeaveHandler(leaveService services.ILeaveService) *LeaveHandler {
	return &LeaveHandler{
		leaveService: leaveService,
	}
}

func (h *LeaveHandler) RegisterRoutes(r *mux.Router) {
	// Leave Type Routes
	r.HandleFunc("/companies/{company_id}/leave-types", h.CreateLeaveType).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/leave-types", h.GetLeaveTypeList).Methods(http.MethodGet)
	r.HandleFunc("/leave-types/{id}", h.GetLeaveTypeByID).Methods(http.MethodGet)
	r.HandleFunc("/leave-types/{id}", h.UpdateLeaveType).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/leave-types/{id}", h.DeleteLeaveType).Methods(http.MethodDelete)

	// Leave Request Routes
	r.HandleFunc("/leave-requests", h.RequestLeave).Methods(http.MethodPost)
	r.HandleFunc("/leave-requests", h.GetLeaveRequests).Methods(http.MethodGet)
	r.HandleFunc("/leave-requests/{id}", h.GetLeaveRequest).Methods(http.MethodGet)
	r.HandleFunc("/leave-requests/{id}/approve", h.ApproveLeaveRequest).Methods(http.MethodPost)
	r.HandleFunc("/leave-requests/{id}/reject", h.RejectLeaveRequest).Methods(http.MethodPost)
	r.HandleFunc("/leave-requests/{id}/withdraw", h.WithdrawLeaveRequest).Methods(http.MethodPost)

	// Leave Balance Routes
	r.HandleFunc("/leave-balance", h.CheckBalance).Methods(http.MethodGet)
	r.HandleFunc("/employees/{employee_id}/leave-balances", h.GetEmployeeLeaveBalances).Methods(http.MethodGet)
	r.HandleFunc("/leave-balance/{leave_type_id}", h.GetLeaveBalance).Methods(http.MethodGet)
}

// CreateLeaveType godoc
// @Summary Create leave type
// @Description Create a new leave type (Admin only).
// @Tags leave-types
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateLeaveTypeRequest true "Create leave type payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=models.LeaveType}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/leave-types [post]
func (h *LeaveHandler) CreateLeaveType(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateLeaveTypeRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.leaveService.CreateLeaveType(r.Context(), companyID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLeaveTypeByID godoc
// @Summary Get leave type by ID
// @Description Retrieve a leave type by ID.
// @Tags leave-types
// @Produce json
// @Param id path string true "Leave Type ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-types/{id} [get]
func (h *LeaveHandler) GetLeaveTypeByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleEmployee, roleManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	leaveTypeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid leave type id")
		return
	}

	result, err := h.leaveService.GetLeaveTypeByID(r.Context(), leaveTypeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLeaveTypeList godoc
// @Summary List leave types
// @Description Get paginated list of leave types for a company.
// @Tags leave-types
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param status query string false "Leave type status (active/inactive)"
// @Param search query string false "Search by name or code"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/leave-types [get]
func (h *LeaveHandler) GetLeaveTypeList(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleEmployee, roleManager)
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

	pageSize, err := parsePositiveInt(r.URL.Query().Get("pageSize"), 10)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid pageSize")
		return
	}

	req := &dto.LeaveTypeListRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   r.URL.Query().Get("status"),
		Search:   r.URL.Query().Get("search"),
	}

	result, err := h.leaveService.GetLeaveTypeList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateLeaveType godoc
// @Summary Update leave type
// @Description Update a leave type by ID (Admin only).
// @Tags leave-types
// @Accept json
// @Produce json
// @Param id path string true "Leave Type ID"
// @Param request body dto.UpdateLeaveTypeRequest true "Update leave type payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-types/{id} [put]
// @Router /leave-types/{id} [patch]
func (h *LeaveHandler) UpdateLeaveType(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	leaveTypeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid leave type id")
		return
	}

	var req dto.UpdateLeaveTypeRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.leaveService.UpdateLeaveType(r.Context(), leaveTypeID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// DeleteLeaveType godoc
// @Summary Delete leave type
// @Description Delete a leave type by ID (Admin only).
// @Tags leave-types
// @Produce json
// @Param id path string true "Leave Type ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-types/{id} [delete]
func (h *LeaveHandler) DeleteLeaveType(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	leaveTypeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid leave type id")
		return
	}

	if err := h.leaveService.DeleteLeaveType(r.Context(), leaveTypeID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "leave type deleted",
	})
}

// RequestLeave godoc
// @Summary Request leave
// @Description Employee requests leave.
// @Tags leave-requests
// @Accept json
// @Produce json
// @Param request body dto.RequestLeaveRequest true "Request leave payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=models.LeaveRequest}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests [post]
func (h *LeaveHandler) RequestLeave(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	var req dto.RequestLeaveRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.leaveService.RequestLeave(r.Context(), employeeID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLeaveRequest godoc
// @Summary Get leave request
// @Description Retrieve a leave request by ID.
// @Tags leave-requests
// @Produce json
// @Param id path string true "Leave Request ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.LeaveRequest}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests/{id} [get]
func (h *LeaveHandler) GetLeaveRequest(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleEmployee, roleManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	result, err := h.leaveService.GetLeaveRequest(r.Context(), requestID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLeaveRequests godoc
// @Summary List leave requests
// @Description Get paginated list of leave requests.
// @Tags leave-requests
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param employeeId query string false "Filter by employee ID"
// @Param status query string false "Filter by status (pending/approved/rejected)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests [get]
func (h *LeaveHandler) GetLeaveRequests(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleEmployee, roleManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
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

	var employeeID uuid.UUID
	if empIDStr := r.URL.Query().Get("employeeId"); empIDStr != "" {
		parsed, err := uuid.Parse(empIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid employeeId")
			return
		}
		employeeID = parsed
	}

	req := &dto.LeaveRequestListRequest{
		Page:       page,
		PageSize:   pageSize,
		EmployeeID: employeeID,
		Status:     r.URL.Query().Get("status"),
	}

	result, err := h.leaveService.GetLeaveRequests(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// ApproveLeaveRequest godoc
// @Summary Approve leave request
// @Description Manager approves a pending leave request.
// @Tags leave-requests
// @Accept json
// @Produce json
// @Param id path string true "Leave Request ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.LeaveRequest}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests/{id}/approve [post]
func (h *LeaveHandler) ApproveLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	approverID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	result, err := h.leaveService.ApproveLeaveRequest(r.Context(), requestID, approverID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// RejectLeaveRequest godoc
// @Summary Reject leave request
// @Description Manager rejects a pending leave request.
// @Tags leave-requests
// @Accept json
// @Produce json
// @Param id path string true "Leave Request ID"
// @Param request body dto.RejectLeaveRequestRequest true "Rejection payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.LeaveRequest}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests/{id}/reject [post]
func (h *LeaveHandler) RejectLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req dto.RejectLeaveRequestRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rejectorID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	result, err := h.leaveService.RejectLeaveRequest(r.Context(), requestID, rejectorID, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// WithdrawLeaveRequest godoc
// @Summary Withdraw leave request
// @Description Employee withdraws their pending leave request.
// @Tags leave-requests
// @Produce json
// @Param id path string true "Leave Request ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.LeaveRequest}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-requests/{id}/withdraw [post]
func (h *LeaveHandler) WithdrawLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	result, err := h.leaveService.WithdrawLeaveRequest(r.Context(), requestID, employeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// CheckBalance godoc
// @Summary Check available balance
// @Description Check available leave days for an employee.
// @Tags leave-balance
// @Produce json
// @Param leaveTypeId query string true "Leave Type ID"
// @Param year query int true "Year"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=float64}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-balance [get]
func (h *LeaveHandler) CheckBalance(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	leaveTypeIDStr := r.URL.Query().Get("leaveTypeId")
	if leaveTypeIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "leaveTypeId is required")
		return
	}

	leaveTypeID, err := uuid.Parse(leaveTypeIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid leaveTypeId")
		return
	}

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "year is required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2100 {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid year")
		return
	}

	available, err := h.leaveService.CheckAvailableBalance(r.Context(), employeeID, leaveTypeID, year)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    available,
	})
}

// GetEmployeeLeaveBalances godoc
// @Summary Get employee leave balances
// @Description Get all leave balances for an employee in a specific year.
// @Tags leave-balance
// @Produce json
// @Param employee_id path string true "Employee ID"
// @Param year query int true "Year"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]models.LeaveBalance}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /employees/{employee_id}/leave-balances [get]
func (h *LeaveHandler) GetEmployeeLeaveBalances(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeToken(r, roleSuperAdmin, roleHRManager, roleEmployee, roleManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	empIDStr := mux.Vars(r)["employee_id"]
	employeeID, err := uuid.Parse(empIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "year is required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2100 {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid year")
		return
	}

	result, err := h.leaveService.GetEmployeeLeaveBalances(r.Context(), employeeID, year)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetLeaveBalance godoc
// @Summary Get specific leave balance
// @Description Get a specific leave balance for an employee.
// @Tags leave-balance
// @Produce json
// @Param leave_type_id path string true "Leave Type ID"
// @Param year query int true "Year"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=models.LeaveBalance}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /leave-balance/{leave_type_id} [get]
func (h *LeaveHandler) GetLeaveBalance(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeToken(r, roleEmployee, roleManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id in token")
		return
	}

	leaveTypeIDStr := mux.Vars(r)["leave_type_id"]
	leaveTypeID, err := uuid.Parse(leaveTypeIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid leave type id")
		return
	}

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "year is required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2100 {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid year")
		return
	}

	result, err := h.leaveService.GetLeaveBalance(r.Context(), employeeID, leaveTypeID, year)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}
