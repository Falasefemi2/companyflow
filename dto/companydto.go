package dto

import (
	"time"

	"github.com/falasefemi2/companyflowlow/utils"
)

type CreateCompanyRequest struct {
	Name               string               `json:"name" validate:"required,min=2,max=255"`
	Slug               string               `json:"slug" validate:"required,min=2,max=255"`
	Industry           string               `json:"industry" validate:"omitempty,max=100"`
	Country            string               `json:"country" validate:"omitempty,max=100"`
	Timezone           string               `json:"timezone" validate:"omitempty,max=50"`
	Currency           string               `json:"currency" validate:"omitempty,max=10"`
	RegistrationNumber string               `json:"registration_number" validate:"omitempty,max=100"`
	TaxID              string               `json:"tax_id" validate:"omitempty,max=100"`
	Address            string               `json:"address" validate:"omitempty"`
	Phone              string               `json:"phone" validate:"omitempty,max=100"`
	LogoURL            string               `json:"logo_url" validate:"omitempty"`
	Status             string               `json:"status" validate:"omitempty,oneof=active suspended inactive"`
	Admin              *CompanyAdminRequest `json:"admin" validate:"required"`
}

type CompanyAdminRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone" validate:"omitempty"`
}

type UpdateCompanyRequest struct {
	Name               *string `json:"name" validate:"omitempty,min=2,max=255"`
	Slug               *string `json:"slug" validate:"omitempty,min=2,max=255"`
	Industry           *string `json:"industry" validate:"omitempty,max=100"`
	Country            *string `json:"country" validate:"omitempty,max=100"`
	Timezone           *string `json:"timezone" validate:"omitempty,max=50"`
	Currency           *string `json:"currency" validate:"omitempty,max=10"`
	RegistrationNumber *string `json:"registration_number" validate:"omitempty,max=100"`
	TaxID              *string `json:"tax_id" validate:"omitempty,max=100"`
	Address            *string `json:"address" validate:"omitempty"`
	Phone              *string `json:"phone" validate:"omitempty,max=100"`
	LogoURL            *string `json:"logo_url" validate:"omitempty"`
	Status             *string `json:"status" validate:"omitempty,oneof=active suspended inactive"`
}

type CompanyResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	Industry           string    `json:"industry"`
	Country            string    `json:"country"`
	Timezone           string    `json:"timezone"`
	Currency           string    `json:"currency"`
	RegistrationNumber string    `json:"registration_number"`
	TaxID              string    `json:"tax_id"`
	Address            string    `json:"address"`
	Phone              string    `json:"phone"`
	LogoURL            string    `json:"logo_url"`
	Status             string    `json:"status"`
	Settings           []byte    `json:"settings"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CompanyListRequest struct {
	utils.PaginationParams
	Status string `json:"status" validate:"omitempty,oneof=active suspended inactive"`
	Search string `json:"search" validate:"omitempty"` // Search by name or slug
}
