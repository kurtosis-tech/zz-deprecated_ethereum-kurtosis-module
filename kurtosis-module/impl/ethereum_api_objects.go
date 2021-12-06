package impl

import "github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"

// Struct representing object that will come back from the Ethereum cluster when getting node info
type EthAPINodeInfoResponse struct {
	Result EthAPINodeInfo `json:"result"`
}

type EthAPINodeInfo struct {
	ServiceID services.ServiceID `json:"service_id"`
	Enode     string             `json:"enode"`
}

type EthAPIAddPeerResponse struct {
	Result bool `json:"result"`
}
