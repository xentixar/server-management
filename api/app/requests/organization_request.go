package requests

type OrganizationRequest struct {
	Name string `form:"name" json:"name" binding:"required,min=5"`
}
