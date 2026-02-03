package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/utils"
)

type CompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{
		pool: pool,
	}
}

func (c *CompanyRepository) CreateCompany(ctx context.Context, company *models.Company) (*models.Company, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		INSERT INTO companies (
			name, slug, industry, country, timezone, currency,
			registration_number, tax_id, address, phone, logo_url,
			status, settings
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
		)
		RETURNING id, name, slug, industry, country, timezone, currency,
				  registration_number, tax_id, address, phone, logo_url,
				  status, settings, created_at, updated_at
	`

	err := c.pool.QueryRow(ctx, query,
		company.Name,
		company.Slug,
		company.Industry,
		company.Country,
		company.Timezone,
		company.Currency,
		company.RegistrationNumber,
		company.TaxID,
		company.Address,
		company.Phone,
		company.LogoURL,
		company.Status,
		company.Settings,
	).Scan(
		&company.ID,
		&company.Name,
		&company.Slug,
		&company.Industry,
		&company.Country,
		&company.Timezone,
		&company.Currency,
		&company.RegistrationNumber,
		&company.TaxID,
		&company.Address,
		&company.Phone,
		&company.LogoURL,
		&company.Status,
		&company.Settings,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return company, nil
}

func (c *CompanyRepository) GetCompanyByID(ctx context.Context, companyID uuid.UUID) (*models.Company, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		SELECT
			id, name, slug, industry, country, timezone, currency,
			registration_number, tax_id, address, phone, logo_url,
			status, settings, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var company models.Company

	err := c.pool.QueryRow(ctx, query, companyID).Scan(
		&company.ID,
		&company.Name,
		&company.Slug,
		&company.Industry,
		&company.Country,
		&company.Timezone,
		&company.Currency,
		&company.RegistrationNumber,
		&company.TaxID,
		&company.Address,
		&company.Phone,
		&company.LogoURL,
		&company.Status,
		&company.Settings,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (c *CompanyRepository) GetCompanyList(
	ctx context.Context,
	listRequest *dto.CompanyListRequest,
) (*utils.PaginatedResponse[*models.Company], error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	where := "WHERE 1=1"
	args := []any{}
	i := 1

	if listRequest.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, listRequest.Status)
		i++
	}

	if listRequest.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", i, i+1)
		search := "%" + listRequest.Search + "%"
		args = append(args, search, search)
		i += 2
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM companies %s", where)
	if err := c.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (listRequest.Page - 1) * listRequest.PageSize

	query := fmt.Sprintf(`
		SELECT
			id, name, slug, industry, country, timezone, currency,
			registration_number, tax_id, address, phone, logo_url,
			status, settings, created_at, updated_at
		FROM companies
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	args = append(args, listRequest.PageSize, offset)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []*models.Company

	for rows.Next() {
		var company models.Company
		if err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.Slug,
			&company.Industry,
			&company.Country,
			&company.Timezone,
			&company.Currency,
			&company.RegistrationNumber,
			&company.TaxID,
			&company.Address,
			&company.Phone,
			&company.LogoURL,
			&company.Status,
			&company.Settings,
			&company.CreatedAt,
			&company.UpdatedAt,
		); err != nil {
			return nil, err
		}
		companies = append(companies, &company)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int((total + int64(listRequest.PageSize) - 1) / int64(listRequest.PageSize))

	return &utils.PaginatedResponse[*models.Company]{
		Data:       companies,
		Total:      total,
		Page:       listRequest.Page,
		PageSize:   listRequest.PageSize,
		TotalPages: totalPages,
		HasNext:    listRequest.Page < totalPages,
		HasPrev:    listRequest.Page > 1,
	}, nil
}

func (c *CompanyRepository) UpdateCompany(ctx context.Context, companyID uuid.UUID, company *models.Company) (*models.Company, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	query := `
		UPDATE companies
		SET
			name = COALESCE(NULLIF($1, ''), name),
			slug = COALESCE(NULLIF($2, ''), slug),
			industry = COALESCE(NULLIF($3, ''), industry),
			country = COALESCE(NULLIF($4, ''), country),
			timezone = COALESCE(NULLIF($5, ''), timezone),
			currency = COALESCE(NULLIF($6, ''), currency),
			registration_number = COALESCE(NULLIF($7, ''), registration_number),
			tax_id = COALESCE(NULLIF($8, ''), tax_id),
			address = COALESCE(NULLIF($9, ''), address),
			phone = COALESCE(NULLIF($10, ''), phone),
			logo_url = COALESCE(NULLIF($11, ''), logo_url),
			status = COALESCE(NULLIF($12, ''), status),
			settings = CASE WHEN $13::jsonb IS DISTINCT FROM NULL THEN $13::jsonb ELSE settings END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
		RETURNING id, name, slug, industry, country, timezone, currency,
				  registration_number, tax_id, address, phone, logo_url,
				  status, settings, created_at, updated_at
	`

	var updated models.Company
	err := c.pool.QueryRow(ctx, query,
		company.Name,
		company.Slug,
		company.Industry,
		company.Country,
		company.Timezone,
		company.Currency,
		company.RegistrationNumber,
		company.TaxID,
		company.Address,
		company.Phone,
		company.LogoURL,
		company.Status,
		company.Settings,
		companyID,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Slug,
		&updated.Industry,
		&updated.Country,
		&updated.Timezone,
		&updated.Currency,
		&updated.RegistrationNumber,
		&updated.TaxID,
		&updated.Address,
		&updated.Phone,
		&updated.LogoURL,
		&updated.Status,
		&updated.Settings,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (c *CompanyRepository) DeleteCompany(ctx context.Context, companyID uuid.UUID, softDelete bool) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if softDelete {
		result, err := c.pool.Exec(
			ctx,
			"UPDATE companies SET status = 'inactive', updated_at = CURRENT_TIMESTAMP WHERE id = $1",
			companyID,
		)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("company not found")
		}
		return nil
	}

	result, err := c.pool.Exec(ctx, "DELETE FROM companies WHERE id = $1", companyID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("company not found")
	}

	return nil
}
