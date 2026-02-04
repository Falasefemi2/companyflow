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

type EmployeeHandler struct {
	employeeService services.IEmployeeService
}

func NewEmployeeHandler(employeeService services.IEmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: employeeService}
}

func (h *EmployeeHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/employees", h.CreateEmployee).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/employees", h.GetEmployeeList).Methods(http.MethodGet)
	r.HandleFunc("/employees/{id}", h.GetEmployeeByID).Methods(http.MethodGet)
	r.HandleFunc("/employees/{id}", h.UpdateEmployee).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/employees/{id}", h.DeleteEmployee).Methods(http.MethodDelete)
}

// CreateEmployee godoc
// @Summary Create employee
// @Description Create a new employee.
// @Tags employees
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateEmployeeRequest true "Create employee payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=dto.EmployeeResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/employees [post]
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateEmployeeRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.employeeService.CreateEmployee(r.Context(), companyID, claims.Role, &req)
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

// GetEmployeeByID godoc
// @Summary Get employee by ID
// @Description Retrieve an employee by ID.
// @Tags employees
// @Produce json
// @Param id path string true "Employee ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.EmployeeResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /employees/{id} [get]
func (h *EmployeeHandler) GetEmployeeByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	employeeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	result, err := h.employeeService.GetEmployeeByID(r.Context(), employeeID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetEmployeeList godoc
// @Summary List employees
// @Description Get paginated list of employees.
// @Tags employees
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Employee status"
// @Param department_id query string false "Department ID"
// @Param manager_id query string false "Manager ID"
// @Param employment_type query string false "Employment type"
// @Param search query string false "Search by name or email"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=PaginatedEmployeeResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/employees [get]
func (h *EmployeeHandler) GetEmployeeList(w http.ResponseWriter, r *http.Request) {
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

	managerID := r.URL.Query().Get("manager_id")
	if managerID != "" {
		if _, err := uuid.Parse(managerID); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid manager_id")
			return
		}
	}

	req := &dto.EmployeeListRequest{
		PaginationParams: utils.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		Status:         r.URL.Query().Get("status"),
		DepartmentID:   departmentID,
		ManagerID:      managerID,
		EmploymentType: r.URL.Query().Get("employment_type"),
		Search:         r.URL.Query().Get("search"),
	}

	result, err := h.employeeService.GetEmployeeList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateEmployee godoc
// @Summary Update employee
// @Description Update an employee by ID.
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param request body dto.UpdateEmployeeRequest true "Update employee payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.EmployeeResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /employees/{id} [put]
// @Router /employees/{id} [patch]
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	employeeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.employeeService.UpdateEmployee(r.Context(), employeeID, &req)
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

// DeleteEmployee godoc
// @Summary Delete employee
// @Description Delete an employee by ID (soft delete unless hard_delete=true).
// @Tags employees
// @Produce json
// @Param id path string true "Employee ID"
// @Param hard_delete query bool false "Hard delete"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	employeeID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	hardDelete := r.URL.Query().Get("hard_delete") == "true"

	if err := h.employeeService.DeleteEmployee(r.Context(), employeeID.String(), hardDelete); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "employee deleted",
	})
}

func parseEmployeeCompanyID(r *http.Request, claims *utils.AuthClaims) (uuid.UUID, error) {
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

func authorizeEmployeeToken(r *http.Request, allowedRoles ...string) (*utils.AuthClaims, error) {
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
