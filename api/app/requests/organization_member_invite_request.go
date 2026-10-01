package requests

type OrganizationMemberInviteRequest struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Role  string `form:"role" json:"role" binding:"required,role"`
}
