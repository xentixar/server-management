package constants

const AgentProtocolVersion = 1

const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusSucceeded = "succeeded"
	JobStatusFailed    = "failed"

	JobKindServerProvision    = "server.provision"
	JobKindRuntimePHPInstall  = "runtime.php.install"
	JobKindRuntimeNodeInstall = "runtime.node.install"
	JobKindRuntimePHPDefault  = "runtime.php.default"
	JobKindRuntimeNodeDefault = "runtime.node.default"
	JobKindRuntimePHPRemove   = "runtime.php.remove"
	JobKindRuntimeNodeRemove  = "runtime.node.remove"
)
