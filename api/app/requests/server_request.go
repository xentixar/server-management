package requests

type ServerSSHFields struct {
	IPAddress            string `json:"ip" binding:"required"`
	SSHPort              int    `json:"ssh_port" binding:"omitempty,min=1,max=65535"`
	SSHUser              string `json:"ssh_user" binding:"required"`
	AuthenticationMethod string `json:"auth_method" binding:"required,oneof=key password"`
	PrivateKey           string `json:"private_key"`
	Passphrase           string `json:"passphrase"`
	Password             string `json:"password"`
}

type TestConnectionRequest struct {
	ServerSSHFields
}

type CreateServerRequest struct {
	Hostname string `json:"hostname" binding:"required,min=2"`
	ServerSSHFields
	Region string   `json:"region"`
	Labels []string `json:"labels"`
	Notes  string   `json:"notes"`
}

type UpdateServerRequest struct {
	Hostname             string   `json:"hostname" binding:"required,min=2"`
	IPAddress            string   `json:"ip" binding:"required"`
	SSHPort              int      `json:"ssh_port" binding:"omitempty,min=1,max=65535"`
	SSHUser              string   `json:"ssh_user" binding:"required"`
	AuthenticationMethod string   `json:"auth_method" binding:"required,oneof=key password"`
	PrivateKey           string   `json:"private_key"`
	Passphrase           string   `json:"passphrase"`
	Password             string   `json:"password"`
	Region               string   `json:"region"`
	Timezone             string   `json:"timezone"`
	Labels               []string `json:"labels"`
	Notes                string   `json:"notes"`
}
