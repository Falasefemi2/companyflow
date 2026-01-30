package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/falasefemi2/companyflowlow/models"
)

type DesignationRepository struct {
	pool *pgxpool.Pool
}

func NewDesignationRepository(pool *pgxpool.Pool) *DesignationRepository {
	return &DesignationRepository{
		pool: pool,
	}
}

func (d *DesignationRepository) CreateDesignation(ctx context.Context, designation *models.Designation) (*models.Designation, error) {
	return nil, nil
}
