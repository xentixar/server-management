package requests

type UpdateMemberRoleRequest struct {
	Role string `form:"role" json:"role" binding:"required,role"`
}
