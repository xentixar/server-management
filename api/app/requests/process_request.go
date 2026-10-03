package requests

type CreateServerServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Command     string `json:"command" binding:"required"`
	Directory   string `json:"directory"`
	User        string `json:"user"`
	Processes   int    `json:"processes"`
	AutoRestart *bool  `json:"auto_restart"`
	Autostart   *bool  `json:"autostart"`
	Enabled     *bool  `json:"enabled"`
}

type UpdateServerServiceRequest struct {
	Command     *string `json:"command"`
	Directory   *string `json:"directory"`
	User        *string `json:"user"`
	Processes   *int    `json:"processes"`
	AutoRestart *bool   `json:"auto_restart"`
	Autostart   *bool   `json:"autostart"`
	Enabled     *bool   `json:"enabled"`
}

type ControlServerServiceRequest struct {
	Action string `json:"action" binding:"required"`
}

type CreateServerCronRequest struct {
	Schedule string `json:"schedule" binding:"required"`
	Command  string `json:"command" binding:"required"`
	User     string `json:"user"`
	Timezone string `json:"timezone"`
	Enabled  *bool  `json:"enabled"`
}

type UpdateServerCronRequest struct {
	Schedule *string `json:"schedule"`
	Command  *string `json:"command"`
	User     *string `json:"user"`
	Timezone *string `json:"timezone"`
	Enabled  *bool   `json:"enabled"`
}

type CreateServerFirewallRuleRequest struct {
	Name      string `json:"name"`
	Action    string `json:"action" binding:"required"`
	Direction string `json:"direction"`
	Protocol  string `json:"protocol" binding:"required"`
	PortStart *int   `json:"port_start"`
	PortEnd   *int   `json:"port_end"`
	Source    string `json:"source"`
	Position  *int   `json:"position"`
	Enabled   *bool  `json:"enabled"`
}

type UpdateServerFirewallRuleRequest struct {
	Name      *string `json:"name"`
	Action    *string `json:"action"`
	Direction *string `json:"direction"`
	Protocol  *string `json:"protocol"`
	PortStart *int    `json:"port_start"`
	PortEnd   *int    `json:"port_end"`
	Source    *string `json:"source"`
	Position  *int    `json:"position"`
	Enabled   *bool   `json:"enabled"`
}

type CreateServerSshKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type UpdateServerSshKeyRequest struct {
	Name      *string `json:"name"`
	PublicKey *string `json:"public_key"`
}
