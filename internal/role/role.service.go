package role

import dto "go-app/internal/dto"

type IRoleService interface {
	CreateRole(createDto dto.CreateRoleDto) *dto.RoleResponseDto
	GetAllRole() *map[string]dto.RoleDto
}

type roleService struct {
	roleRepo IRoleRepo
}

func (r *roleService) CreateRole(createDto dto.CreateRoleDto) *dto.RoleResponseDto {
	result := r.roleRepo.CreateRole(createDto)
	if result == nil {
		return nil
	}
	return &dto.RoleResponseDto{
		RoleId:   result.ID.Hex(),
		RoleName: result.RoleName,
		RoleNote: result.RoleNote,
	}
}

// GetAllRole implements [IRoleService].
func (r *roleService) GetAllRole() *map[string]dto.RoleDto {
	result := r.roleRepo.GetAllRole() // trả về địa chỉ
	return result                     // trả về giá trị của địa chỉ trong result
}

func NewRoleService(roleRepo IRoleRepo) IRoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}
