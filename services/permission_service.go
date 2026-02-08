package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/utils"
)

type IPermissionService interface {
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissions []*models.Permission) (*dto.RoleResponse, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*dto.PermissionResponse, error)
	AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission *models.Permission) (*dto.RoleResponse, error)
	RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) (*dto.RoleResponse, error)
}

type PermissionService struct {
	roleRepo *repositories.RoleRepository
}

func NewPermissionService(roleRepo *repositories.RoleRepository) *PermissionService {
	return &PermissionService{roleRepo: roleRepo}
}

func (ps *PermissionService) SetRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
	permissions []*models.Permission,
) (*dto.RoleResponse, error) {
	if permissions == nil {
		permissions = []*models.Permission{}
	}

	if err := ps.validatePermissions(permissions); err != nil {
		return nil, &utils.ValidationError{Message: err.Error()}
	}

	updatedRole, err := ps.roleRepo.SetRolePermissions(ctx, roleID, permissions)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updatedRole), nil
}

func (ps *PermissionService) GetRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
) ([]*dto.PermissionResponse, error) {
	permissions, err := ps.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	var responses []*dto.PermissionResponse
	for _, perm := range permissions {
		responses = append(responses, toPermissionResponse(perm))
	}

	return responses, nil
}

func (ps *PermissionService) AddPermissionToRole(
	ctx context.Context,
	roleID uuid.UUID,
	permission *models.Permission,
) (*dto.RoleResponse, error) {
	if permission == nil {
		return nil, &utils.ValidationError{Message: "permission cannot be nil"}
	}

	if err := ps.validatePermission(permission); err != nil {
		return nil, &utils.ValidationError{Message: err.Error()}
	}

	updatedRole, err := ps.roleRepo.AddPermissionToRole(ctx, roleID, permission)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updatedRole), nil
}

func (ps *PermissionService) RemovePermissionFromRole(
	ctx context.Context,
	roleID uuid.UUID,
	permissionID uuid.UUID,
) (*dto.RoleResponse, error) {
	updatedRole, err := ps.roleRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updatedRole), nil
}

func (ps *PermissionService) validatePermissions(permissions []*models.Permission) error {
	validActions := map[string]bool{
		"create":  true,
		"read":    true,
		"update":  true,
		"delete":  true,
		"approve": true,
		"reject":  true,
		"manage":  true,
	}

	for _, perm := range permissions {
		if !validActions[perm.Action] {
			return fmt.Errorf("invalid action: %s", perm.Action)
		}
		if perm.Resource == "" {
			return fmt.Errorf("resource cannot be empty")
		}
	}

	return nil
}

func (ps *PermissionService) validatePermission(permission *models.Permission) error {
	validActions := map[string]bool{
		"create":  true,
		"read":    true,
		"update":  true,
		"delete":  true,
		"approve": true,
		"reject":  true,
		"manage":  true,
	}

	if !validActions[permission.Action] {
		return fmt.Errorf("invalid action: %s", permission.Action)
	}
	if permission.Resource == "" {
		return fmt.Errorf("resource cannot be empty")
	}

	return nil
}

func toPermissionResponse(perm *models.Permission) *dto.PermissionResponse {
	if perm == nil {
		return nil
	}
	return &dto.PermissionResponse{
		ID:         perm.ID.String(),
		RoleID:     perm.RoleID.String(),
		Action:     perm.Action,
		Resource:   perm.Resource,
		Conditions: perm.Conditions,
		CreatedAt:  perm.CreatedAt,
	}
}
