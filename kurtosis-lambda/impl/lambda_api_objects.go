package impl

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
)

// Struct representing the params that the Lambda can be started with
type LambdaInitArgs struct {
	// Indicates the log level for this Kurtosis Lambda implementation
	LogLevel string `json:"logLevel"`
}

// Struct representing the params that the Lambda will accept when being executed
type LambdaAPIExecuteArgs struct {}

// Struct representing the result that will be returned to the user on execute
type LambdaAPIExecuteResult struct {
	BootnodeServiceID     services.ServiceID                       `json:"bootnode_service_id"`
	NodeInfo			  map[services.ServiceID]*LambdaAPINodeInfo `json:"node_info"`
	SignerKeystoreContent string                                   `json:"signer_keystore_content"`
	SignerAccountPassword string                                   `json:"signer_account_password"`
}

type LambdaAPINodeInfo struct {
	IPAddrInsideNetwork string 			`json:"ip_addr_inside_network"`
	ExposedPortsSet	map[string]bool		`json:"exposed_ports_set"`
	PortBindingsOnLocalMachine map[string]*kurtosis_core_rpc_api_bindings.PortBinding `json:"port_bindings_on_local_machine"`
}
