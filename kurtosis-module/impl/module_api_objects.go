package impl

import "github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"

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
	IPAddrOnHostMachine string 			`json:"ip_addr_on_host_machine"`
	RpcPortOnHostMachine uint16			`json:"rpcPortOnHostMachine"`
	WsPortOnHostMachine uint16			`json:"wsPortOnHostMachine"`
	TcpDiscoveryPortOnHostMachine uint16			`json:"tcpDiscoveryPortOnHostMachine"`
	UdpDiscoveryPortOnHostMachine uint16			`json:"udpDiscoveryPortOnHostMachine"`
}
