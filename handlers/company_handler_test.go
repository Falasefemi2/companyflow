package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/utils"
)

type mockCompanyService struct {
	createFn   func(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error)
	getByIDFn  func(ctx context.Context, companyID uuid.UUID) (*dto.CompanyResponse, error)
	listFn     func(ctx context.Context, req *dto.CompanyListRequest) (*utils.PaginatedResponse[*models.Company], error)
	updateFn   func(ctx context.Context, companyID uuid.UUID, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error)
	deleteFn   func(ctx context.Context, companyID uuid.UUID, softDelete bool) error
}

func (m *mockCompanyService) CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
	return m.createFn(ctx, req)
}

func (m *mockCompanyService) GetCompanyByID(ctx context.Context, companyID uuid.UUID) (*dto.CompanyResponse, error) {
	return m.getByIDFn(ctx, companyID)
}

func (m *mockCompanyService) GetCompanyList(ctx context.Context, req *dto.CompanyListRequest) (*utils.PaginatedResponse[*models.Company], error) {
	return m.listFn(ctx, req)
}

func (m *mockCompanyService) UpdateCompany(ctx context.Context, companyID uuid.UUID, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	return m.updateFn(ctx, companyID, req)
}

func (m *mockCompanyService) DeleteCompany(ctx context.Context, companyID uuid.UUID, softDelete bool) error {
	return m.deleteFn(ctx, companyID, softDelete)
}

func TestCompanyHandler_CreateCompany(t *testing.T) {
	mockService := &mockCompanyService{
		createFn: func(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
			if req.Name != "Test Company" {
				t.Errorf("expected name Test Company, got %s", req.Name)
			}
			return &dto.CompanyResponse{
				ID:        uuid.New().String(),
				Name:      req.Name,
				Slug:      req.Slug,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	body := `{"name":"Test Company","slug":"test-company"}`
	req := httptest.NewRequest(http.MethodPost, "/companies", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateCompany(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["success"] != true {
		t.Fatalf("expected success true, got %v", payload["success"])
	}
}

func TestCompanyHandler_CreateCompany_InvalidJSON(t *testing.T) {
	mockService := &mockCompanyService{
		createFn: func(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
			return nil, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/companies", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	handler.CreateCompany(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCompanyHandler_GetCompanyByID(t *testing.T) {
	companyID := uuid.New()
	mockService := &mockCompanyService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*dto.CompanyResponse, error) {
			if id != companyID {
				t.Errorf("expected id %s, got %s", companyID, id)
			}
			return &dto.CompanyResponse{
				ID:   id.String(),
				Name: "Company",
				Slug: "company",
			}, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/companies/"+companyID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": companyID.String()})
	rec := httptest.NewRecorder()

	handler.GetCompanyByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCompanyHandler_GetCompanyByID_InvalidID(t *testing.T) {
	mockService := &mockCompanyService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*dto.CompanyResponse, error) {
			return nil, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/companies/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rec := httptest.NewRecorder()

	handler.GetCompanyByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCompanyHandler_GetCompanyList(t *testing.T) {
	mockService := &mockCompanyService{
		listFn: func(ctx context.Context, req *dto.CompanyListRequest) (*utils.PaginatedResponse[*models.Company], error) {
			if req.Page != 1 || req.PageSize != 2 {
				t.Errorf("unexpected pagination: %+v", req.PaginationParams)
			}
			if req.Status != "active" || req.Search != "Acme" {
				t.Errorf("unexpected filters: status=%s search=%s", req.Status, req.Search)
			}
			return &utils.PaginatedResponse[*models.Company]{
				Data:       []*models.Company{},
				Total:      0,
				Page:       1,
				PageSize:   2,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			}, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/companies?page=1&page_size=2&status=active&search=Acme", nil)
	rec := httptest.NewRecorder()

	handler.GetCompanyList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCompanyHandler_UpdateCompany(t *testing.T) {
	companyID := uuid.New()
	mockService := &mockCompanyService{
		updateFn: func(ctx context.Context, id uuid.UUID, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
			if id != companyID {
				t.Errorf("expected id %s, got %s", companyID, id)
			}
			if req.Name == nil || *req.Name != "Updated Company" {
				t.Errorf("expected name Updated Company, got %+v", req.Name)
			}
			return &dto.CompanyResponse{
				ID:   id.String(),
				Name: *req.Name,
				Slug: "updated-company",
			}, nil
		},
	}

	handler := NewCompanyHandler(mockService)

	body := `{"name":"Updated Company"}`
	req := httptest.NewRequest(http.MethodPatch, "/companies/"+companyID.String(), bytes.NewBufferString(body))
	req = mux.SetURLVars(req, map[string]string{"id": companyID.String()})
	rec := httptest.NewRecorder()

	handler.UpdateCompany(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCompanyHandler_DeleteCompany(t *testing.T) {
	companyID := uuid.New()
	mockService := &mockCompanyService{
		deleteFn: func(ctx context.Context, id uuid.UUID, softDelete bool) error {
			if id != companyID {
				t.Errorf("expected id %s, got %s", companyID, id)
			}
			if softDelete {
				t.Error("expected hard delete")
			}
			return nil
		},
	}

	handler := NewCompanyHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/companies/"+companyID.String()+"?hard_delete=true", nil)
	req = mux.SetURLVars(req, map[string]string{"id": companyID.String()})
	rec := httptest.NewRecorder()

	handler.DeleteCompany(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
