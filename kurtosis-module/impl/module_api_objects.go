package impl

import "github.com/kurtosis-tech/kurtosis-sdk/api/golang/core/lib/services"

// Struct representing the params that the module can be started with
type ModuleInitArgs struct {
	// Indicates the log level for this Kurtosis module implementation
	LogLevel string `json:"logLevel"`
}

// Struct representing the params that the executable module will accept when being executed
type ModuleAPIExecuteArgs struct{}

// Struct representing the result that will be returned to the user on execute
type ModuleAPIExecuteResult struct {
	BootnodeServiceID     services.ServiceID                                `json:"bootnode_service_id"`
	NodeInfo              map[services.ServiceID]*ModuleAPIEthereumNodeInfo `json:"node_info"`
	SignerKeystoreContent string                                            `json:"signer_keystore_content"`
	SignerAccountPassword string                                            `json:"signer_account_password"`
}

type ModuleAPIEthereumNodeInfo struct {
	IPAddrInsideNetwork string `json:"ip_addr_inside_network"`
	IPAddrOnHostMachine string `json:"ip_addr_on_host_machine"`
	RpcPortId           string `json:"rpc_port_id"`
	WsPortId            string `json:"ws_port_id"`
	TcpDiscoveryPortId  string `json:"tcp_discovery_port_id"`
	UdpDiscoveryPortId  string `json:"udp_discovery_port_id"`
}
