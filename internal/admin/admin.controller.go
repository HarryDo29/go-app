package admin

type AdminController struct {
	adminService IAdminService
}

func NewAdminController(adminService IAdminService) *AdminController {
	return &AdminController{
		adminService: adminService,
	}
}
