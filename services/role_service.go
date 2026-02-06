package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IRoleService interface {
	CreateRole(ctx context.Context, companyID uuid.UUID, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*dto.RoleResponse, error)
	GetRoleList(ctx context.Context, companyID uuid.UUID, listRequest *dto.RoleListRequest) (*utils.PaginatedResponse[*models.Role], error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
}

type RoleService struct {
	roleRepo *repositories.RoleRepository
}

func NewRoleService(roleRepo *repositories.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (rs *RoleService) CreateRole(
	ctx context.Context,
	companyID uuid.UUID,
	req *dto.CreateRoleRequest,
) (*dto.RoleResponse, error) {
	if req == nil {
		return nil, nil
	}

	description := (*string)(nil)
	if req.Description != "" {
		description = &req.Description
	}

	permissions := req.PermissionsCache
	if permissions == nil {
		permissions = []string{}
	}

	role := &models.Role{
		CompanyID:        &companyID,
		Name:             req.Name,
		Description:      description,
		IsSystemRole:     false,
		PermissionsCache: permissions,
	}

	created, err := rs.roleRepo.CreateRole(ctx, role)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(created), nil
}

func (rs *RoleService) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*dto.RoleResponse, error) {
	role, err := rs.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return toRoleResponse(role), nil
}

func (rs *RoleService) GetRoleList(
	ctx context.Context,
	companyID uuid.UUID,
	listRequest *dto.RoleListRequest,
) (*utils.PaginatedResponse[*models.Role], error) {
	return rs.roleRepo.GetRoleList(ctx, companyID, listRequest)
}

func (rs *RoleService) UpdateRole(
	ctx context.Context,
	roleID uuid.UUID,
	req *dto.UpdateRoleRequest,
) (*dto.RoleResponse, error) {
	if req == nil {
		return nil, nil
	}

	role := &models.Role{}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = req.Description
	}
	if req.PermissionsCache != nil {
		role.PermissionsCache = *req.PermissionsCache
	}

	updated, err := rs.roleRepo.UpdateRole(ctx, roleID, role)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updated), nil
}

func (rs *RoleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	return rs.roleRepo.DeleteRole(ctx, roleID)
}

func toRoleResponse(role *models.Role) *dto.RoleResponse {
	if role == nil {
		return nil
	}

	var companyID *string
	if role.CompanyID != nil {
		value := role.CompanyID.String()
		companyID = &value
	}

	return &dto.RoleResponse{
		ID:               role.ID.String(),
		CompanyID:        companyID,
		Name:             role.Name,
		Description:      role.Description,
		IsSystemRole:     role.IsSystemRole,
		PermissionsCache: role.PermissionsCache,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}
}
