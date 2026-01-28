package services

import (
	"context"

	"github.com/falasefemi2/companyflowlow/dto"
)

type IAuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (string, error)
}
