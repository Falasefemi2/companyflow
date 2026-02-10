package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type BulkEmployeeHandler struct {
	bulkEmployeeService *services.BulkEmployeeService
}

func NewBulkEmployeeHandler(
	bulkEmployeeService *services.BulkEmployeeService,
) *BulkEmployeeHandler {
	return &BulkEmployeeHandler{
		bulkEmployeeService: bulkEmployeeService,
	}
}

func (h *BulkEmployeeHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/employees/bulk", h.UploadBulkEmployees).Methods(http.MethodPost)
}

// UploadBulkEmployees godoc
// @Summary Bulk upload employees
// @Description Upload a CSV file to create employees in bulk.
// @Tags employees
// @Accept multipart/form-data
// @Produce json
// @Param company_id path string true "Company ID"
// @Param file formData file true "CSV file"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=services.BulkCreateEmployeesResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/employees/bulk [post]
func (h *BulkEmployeeHandler) UploadBulkEmployees(w http.ResponseWriter, r *http.Request) {
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

	// Get the file from request
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	// Validate file is CSV
	contentType := strings.ToLower(fileHeader.Header.Get("Content-Type"))
	filename := strings.ToLower(fileHeader.Filename)
	isCSVContentType := strings.Contains(contentType, "csv") || strings.Contains(contentType, "text/plain")
	isCSVExtension := strings.HasSuffix(filename, ".csv")
	if !isCSVContentType && !isCSVExtension {
		utils.RespondWithError(w, http.StatusBadRequest, "file must be a CSV file")
		return
	}

	// Parse CSV
	records, err := parseBulkEmployeeCSV(file)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse CSV: %s", err.Error()))
		return
	}

	if len(records) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "CSV file contains no records")
		return
	}

	// Create service request
	request := &services.BulkCreateEmployeesRequest{
		CompanyID: companyID,
		Records:   records,
	}

	// Call service
	response, err := h.bulkEmployeeService.CreateBulkEmployees(r.Context(), request)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create employees: %s", err.Error()))
		return
	}

	// If there are validation errors, return 400 with details
	if response.FailureCount > 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "validation_failed",
			Data: map[string]interface{}{
				"success_count":     response.SuccessCount,
				"failure_count":     response.FailureCount,
				"validation_errors": response.ValidationErrors,
			},
		})
		return
	}

	// Success - return created employees with temp passwords
	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Employees created successfully. Temporary passwords have been generated.",
		Data:    response,
	})
}

// parseBulkEmployeeCSV parses a CSV file and returns BulkEmployeeRecord slice
// Expected CSV header: email,password,phone,first_name,last_name,date_of_birth,employee_code,
//
//	department_id,designation_id,level_id,role_id,manager_id,status,
//	employment_type,hire_date,gender,address,emergency_contact_name,
//	emergency_contact_phone,profile_image_url
func parseBulkEmployeeCSV(file io.Reader) ([]dto.BulkEmployeeRecord, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create header index map for flexible column ordering
	headerIndex := make(map[string]int)
	for i, header := range headers {
		normalized := normalizeCSVHeader(header)
		headerIndex[normalized] = i
	}

	// Validate required headers exist
	requiredHeaders := []string{
		"email", "first_name", "last_name", "phone", "employee_code",
		"department_id", "designation_id", "level_id", "role_id",
		"status", "employment_type", "hire_date",
	}
	for _, required := range requiredHeaders {
		if _, exists := headerIndex[required]; !exists {
			return nil, fmt.Errorf("missing required CSV column: %s", required)
		}
	}

	var records []dto.BulkEmployeeRecord

	for rowNum := 2; ; rowNum++ { // Start at 2 because row 1 is header
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row %d: %w", rowNum, err)
		}

		record := dto.BulkEmployeeRecord{
			Email:                 getCSVValue(row, headerIndex, "email"),
			Password:              getCSVValue(row, headerIndex, "password"),
			Phone:                 getCSVValue(row, headerIndex, "phone"),
			FirstName:             getCSVValue(row, headerIndex, "first_name"),
			LastName:              getCSVValue(row, headerIndex, "last_name"),
			DateOfBirth:           getCSVValue(row, headerIndex, "date_of_birth"),
			EmployeeCode:          getCSVValue(row, headerIndex, "employee_code"),
			DepartmentID:          getCSVValue(row, headerIndex, "department_id"),
			DesignationID:         getCSVValue(row, headerIndex, "designation_id"),
			LevelID:               getCSVValue(row, headerIndex, "level_id"),
			RoleID:                getCSVValue(row, headerIndex, "role_id"),
			ManagerID:             getCSVValue(row, headerIndex, "manager_id"),
			Status:                getCSVValue(row, headerIndex, "status"),
			EmploymentType:        getCSVValue(row, headerIndex, "employment_type"),
			HireDate:              getCSVValue(row, headerIndex, "hire_date"),
			Gender:                getCSVValue(row, headerIndex, "gender"),
			Address:               getCSVValue(row, headerIndex, "address"),
			EmergencyContactName:  getCSVValue(row, headerIndex, "emergency_contact_name"),
			EmergencyContactPhone: getCSVValue(row, headerIndex, "emergency_contact_phone"),
			ProfileImageUrl:       getCSVValue(row, headerIndex, "profile_image_url"),
		}

		records = append(records, record)
	}

	return records, nil
}

// getCSVValue safely gets a value from a CSV row by header name
func getCSVValue(row []string, headerIndex map[string]int, header string) string {
	normalized := normalizeCSVHeader(header)
	if idx, exists := headerIndex[normalized]; exists && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func normalizeCSVHeader(header string) string {
	trimmed := strings.TrimSpace(header)
	trimmed = strings.TrimPrefix(trimmed, "\ufeff")
	return strings.ToLower(trimmed)
}
