package impl

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
)

// Struct representing the params that the module can be started with
type ModuleInitArgs struct {
	// Indicates the log level for this Kurtosis module implementation
	LogLevel string `json:"logLevel"`
}

// Struct representing the params that the executable module will accept when being executed
type ModuleAPIExecuteArgs struct {}

// Struct representing the result that will be returned to the user on execute
type ModuleAPIExecuteResult struct {
	BootnodeServiceID     services.ServiceID                                `json:"bootnode_service_id"`
	NodeInfo			  map[services.ServiceID]*ModuleAPIEthereumNodeInfo `json:"node_info"`
	SignerKeystoreContent string                                            `json:"signer_keystore_content"`
	SignerAccountPassword string                                            `json:"signer_account_password"`
}

type ModuleAPIEthereumNodeInfo struct {
	IPAddrInsideNetwork string 			`json:"ip_addr_inside_network"`
	ExposedPortsSet	map[string]bool		`json:"exposed_ports_set"`
	PortBindingsOnLocalMachine map[string]*kurtosis_core_rpc_api_bindings.PortBinding `json:"port_bindings_on_local_machine"`
}
