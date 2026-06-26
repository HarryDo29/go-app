package dto

type CreateRoleDto struct {
	RoleName string `json:"role_name"`
	RoleNote string `json:"role_note"`
}

type RoleDto struct {
	RoleId string `json:"role_id"`
	RoleName string `json:"role_name"`
}

type RoleResponseDto struct {
	RoleId string `json:"role_id"`
	RoleName string `json:"role_name"`
	RoleNote string `json:"role_note"`
}
