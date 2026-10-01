package requests

type InviteMemberRequest struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Role  string `form:"role" json:"role" binding:"required,role"`
}
