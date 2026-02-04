package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type ICompanyService interface {
	CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error)
	GetCompanyByID(ctx context.Context, companyID uuid.UUID) (*dto.CompanyResponse, error)
	GetCompanyList(ctx context.Context, listRequest *dto.CompanyListRequest) (*utils.PaginatedResponse[*models.Company], error)
	UpdateCompany(ctx context.Context, companyID uuid.UUID, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error)
	DeleteCompany(ctx context.Context, companyID uuid.UUID, softDelete bool) error
}

type CompanyService struct {
	companyRepo *repositories.CompanyRepository
}

func NewCompanyService(companyRepo *repositories.CompanyRepository) *CompanyService {
	return &CompanyService{
		companyRepo: companyRepo,
	}
}

func (cs *CompanyService) CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
	if req == nil {
		return nil, nil
	}

	if req.Admin == nil {
		return nil, &utils.ValidationError{
			Field:   "admin",
			Message: "admin details are required",
		}
	}

	if req.Admin.Email == "" || req.Admin.Password == "" || req.Admin.FirstName == "" || req.Admin.LastName == "" {
		return nil, &utils.ValidationError{
			Field:   "admin",
			Message: "admin email, password, first_name, and last_name are required",
		}
	}

	if !utils.IsValidEmail(req.Admin.Email) {
		return nil, &utils.ValidationError{
			Field:   "admin.email",
			Message: "admin email is invalid",
		}
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	currency := req.Currency
	if currency == "" {
		currency = "NGN"
	}

	company := &models.Company{
		Name:               req.Name,
		Slug:               req.Slug,
		Industry:           req.Industry,
		Country:            req.Country,
		Timezone:           timezone,
		Currency:           currency,
		RegistrationNumber: req.RegistrationNumber,
		TaxID:              req.TaxID,
		Address:            req.Address,
		Phone:              req.Phone,
		LogoURL:            req.LogoURL,
		Status:             status,
		Settings:           []byte(`{}`),
	}

	hashedPassword, err := utils.HashPassword(req.Admin.Password)
	if err != nil {
		return nil, err
	}

	admin := &models.Employee{
		Email:          req.Admin.Email,
		PasswordHash:   hashedPassword,
		Phone:          req.Admin.Phone,
		FirstName:      req.Admin.FirstName,
		LastName:       req.Admin.LastName,
		Status:         "active",
		EmploymentType: "full_time",
		HireDate:       time.Now().UTC(),
	}

	created, err := cs.companyRepo.CreateCompanyWithAdmin(ctx, company, admin)
	if err != nil {
		return nil, err
	}

	return toCompanyResponse(created), nil
}

func (cs *CompanyService) GetCompanyByID(ctx context.Context, companyID uuid.UUID) (*dto.CompanyResponse, error) {
	company, err := cs.companyRepo.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return toCompanyResponse(company), nil
}

func (cs *CompanyService) GetCompanyList(
	ctx context.Context,
	listRequest *dto.CompanyListRequest,
) (*utils.PaginatedResponse[*models.Company], error) {
	return cs.companyRepo.GetCompanyList(ctx, listRequest)
}

func (cs *CompanyService) UpdateCompany(ctx context.Context, companyID uuid.UUID, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	if req == nil {
		return nil, nil
	}

	company := &models.Company{}

	if req.Name != nil {
		company.Name = *req.Name
	}
	if req.Slug != nil {
		company.Slug = *req.Slug
	}
	if req.Industry != nil {
		company.Industry = *req.Industry
	}
	if req.Country != nil {
		company.Country = *req.Country
	}
	if req.Timezone != nil {
		company.Timezone = *req.Timezone
	}
	if req.Currency != nil {
		company.Currency = *req.Currency
	}
	if req.RegistrationNumber != nil {
		company.RegistrationNumber = *req.RegistrationNumber
	}
	if req.TaxID != nil {
		company.TaxID = *req.TaxID
	}
	if req.Address != nil {
		company.Address = *req.Address
	}
	if req.Phone != nil {
		company.Phone = *req.Phone
	}
	if req.LogoURL != nil {
		company.LogoURL = *req.LogoURL
	}
	if req.Status != nil {
		company.Status = *req.Status
	}

	updated, err := cs.companyRepo.UpdateCompany(ctx, companyID, company)
	if err != nil {
		return nil, err
	}

	return toCompanyResponse(updated), nil
}

func (cs *CompanyService) DeleteCompany(ctx context.Context, companyID uuid.UUID, softDelete bool) error {
	return cs.companyRepo.DeleteCompany(ctx, companyID, softDelete)
}

func toCompanyResponse(company *models.Company) *dto.CompanyResponse {
	if company == nil {
		return nil
	}

	return &dto.CompanyResponse{
		ID:                 company.ID.String(),
		Name:               company.Name,
		Slug:               company.Slug,
		Industry:           company.Industry,
		Country:            company.Country,
		Timezone:           company.Timezone,
		Currency:           company.Currency,
		RegistrationNumber: company.RegistrationNumber,
		TaxID:              company.TaxID,
		Address:            company.Address,
		Phone:              company.Phone,
		LogoURL:            company.LogoURL,
		Status:             company.Status,
		Settings:           company.Settings,
		CreatedAt:          company.CreatedAt,
		UpdatedAt:          company.UpdatedAt,
	}
}
