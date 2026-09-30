package requests

type StoreOrganizationRequest struct {
	Name string `form:"name" json:"name" binding:"required,min=5"`
}
