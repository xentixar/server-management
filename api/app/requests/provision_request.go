package requests

type ProvisionAddons struct {
	Redis    bool `json:"redis"`
	MySQL    bool `json:"mysql"`
	MariaDB  bool `json:"mariadb"`
	Postgres bool `json:"postgres"`
}

type ProvisionServerRequest struct {
	PHPVersions  []string        `json:"php_versions" binding:"required,min=1"`
	NodeVersions []string        `json:"node_versions" binding:"required,min=1"`
	Addons       ProvisionAddons `json:"addons"`
}

type InstallRuntimeRequest struct {
	Runtime    string `json:"runtime" binding:"required,oneof=php node"`
	Version    string `json:"version" binding:"required"`
	SetDefault bool   `json:"set_default"`
}

type RemoveRuntimeRequest struct {
	Runtime string `json:"runtime" binding:"required,oneof=php node"`
	Version string `json:"version" binding:"required"`
}

type SetRuntimeDefaultRequest struct {
	Runtime string `json:"runtime" binding:"required,oneof=php node"`
	Version string `json:"version" binding:"required"`
}
