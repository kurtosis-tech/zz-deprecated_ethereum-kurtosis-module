package impl

import (
	"bytes"
	"encoding/json"
	"fmt"
	static_files_consts "github.com/kurtosis-tech/ethereum-kurtosis-lambda/kurtosis-lambda/static-files-consts"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ethereumDockerImageName = "ethereum/client-go:v1.10.8"

	rpcPort       = 8545
	wsPort        = 8546
	discoveryPort = 30303
	subnetRange   = "/24"

	bootnodeServiceID  = "bootnode"
	ethServiceIdPrefix = "ethereum-node-"
	ethNodesQuantity   = 3

	waitEndpointInitialDelayMilliseconds = 100
	waitEndpointRetries                  = 30
	waitEndpointRetriesDelayMilliseconds = 5

	adminInfoRpcCall           = `{"jsonrpc":"2.0","method": "admin_nodeInfo","params":[],"id":67}`
	execCommandSuccessExitCode = 0
	rpcRequestTimeout          = 30 * time.Second
	jsonContentType            = "application/json"
	enodePrefix                = "enode://"
	handshakeProtocol          = "eth: \"handshake\""
)

type EthereumKurtosisLambda struct {
}

type EthereumKurtosisLambdaParams struct {
}

type EthereumKurtosisLambdaResult struct {
	BootnodeServiceID          services.ServiceID    `json:"bootnode_service_id"`
	NodeServiceIDs             []services.ServiceID  `json:"node_service_ids"`
	GenesisStaticFileID        services.StaticFileID `json:"genesis_static_file_id"`
	PasswordStaticFileID       services.StaticFileID `json:"password_static_file_id"`
	SignerKeystoreStaticFileID services.StaticFileID `json:"signer_keystore_static_file_id"`
}

type AddPeerResponse struct {
	Result bool `json:"result"`
}

type NodeInfoResponse struct {
	Result NodeInfo `json:"result"`
}

type NodeInfo struct {
	ServiceID services.ServiceID `json:"service_id"`
	Enode     string             `json:"enode"`
}

type BootnodeInfo struct {
	ServiceID services.ServiceID `json:"service_id"`
	Enr       string             `json:"enr"`
}

func NewEthereumKurtosisLambda() *EthereumKurtosisLambda {
	return &EthereumKurtosisLambda{}
}

func (e EthereumKurtosisLambda) Execute(networkCtx *networks.NetworkContext, serializedParams string) (serializedResult string, resultError error) {
	logrus.Infof("Ethereum Kurtosis Lambda receives serializedParams '%v'", serializedParams)
	serializedParamsBytes := []byte(serializedParams)
	var params EthereumKurtosisLambdaParams
	if err := json.Unmarshal(serializedParamsBytes, &params); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the Ethereum Kurtosis Lambda serialized params with value '%v'", serializedParams)
	}

	if err := networkCtx.RegisterStaticFiles(static_files_consts.StaticFileFilepaths); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred registering the static files")
	}

	bootnodeInfo, err := startEthBootnode(networkCtx)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the Ethereum Bootnode")
	}

	var nodeServiceIDs []services.ServiceID
	var nodesInfo []*NodeInfo
	for i := 1; i <= ethNodesQuantity; i++ {
		serviceID := services.ServiceID(ethServiceIdPrefix + strconv.Itoa(i))
		nodeInfo, err := starEthNodeByBootnode(networkCtx, serviceID, bootnodeInfo.Enr, nodesInfo)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred starting the Ethereum Node '%v'", serviceID)
		}
		nodesInfo = append(nodesInfo, nodeInfo)
		nodeServiceIDs = append(nodeServiceIDs, nodeInfo.ServiceID)
	}

	ethereumKurtosisLambdaResult := &EthereumKurtosisLambdaResult{
		BootnodeServiceID:          bootnodeInfo.ServiceID,
		NodeServiceIDs:             nodeServiceIDs,
		GenesisStaticFileID:        static_files_consts.GenesisStaticFileID,
		PasswordStaticFileID:       static_files_consts.PasswordStaticFileID,
		SignerKeystoreStaticFileID: static_files_consts.SignerKeystoreFileID,
	}

	result, err := json.Marshal(ethereumKurtosisLambdaResult)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the Ethereum Kurtosis Lambda Result with value '%+v'", ethereumKurtosisLambdaResult)
	}
	stringResult := string(result)

	logrus.Info("Ethereum Kurtosis Lambda executed successfully")
	return stringResult, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func startEthBootnode(networkCtx *networks.NetworkContext) (*BootnodeInfo, error) {
	containerCreationConfig := getContainerCreationConfig()
	runConfigFunc := getEthBootnodeRunConfigFunc()

	serviceCtx, hostPortBindings, err := networkCtx.AddService(bootnodeServiceID, containerCreationConfig, runConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the Ethereum bootnode service")
	}

	if err := networkCtx.WaitForHttpPostEndpointAvailability(bootnodeServiceID, uint32(rpcPort), "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for service with ID '%v' to start", bootnodeServiceID)
	}

	logrus.Infof("Added Ethereum bootnode service with IP: %v and host port bindings: %+v", serviceCtx.GetIPAddress(), hostPortBindings)

	cmd := "geth attach data/geth.ipc --exec admin.nodeInfo.enr"
	exitCode, logOutput, err := serviceCtx.ExecCommand([]string{
		"/bin/sh",
		"-c",
		cmd,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Executing command '%v' returned an error", cmd)
	}
	if exitCode != execCommandSuccessExitCode {
		return nil, stacktrace.NewError("Executing command '%v' returned an failing exit code with logs: %+v", cmd, string(*logOutput))
	}

	bootnodeInfo := &BootnodeInfo{
		ServiceID: serviceCtx.GetServiceID(),
		Enr:       string(*logOutput),
	}

	return bootnodeInfo, nil
}

func starEthNodeByBootnode(networkCtx *networks.NetworkContext, serviceID services.ServiceID, bootnodeEnr string, nodesInfo []*NodeInfo) (*NodeInfo, error) {
	containerCreationConfig := getContainerCreationConfig()
	runConfigFunc := getEthNodeRunConfigFunc(bootnodeEnr)

	serviceCtx, hostPortBindings, err := networkCtx.AddService(serviceID, containerCreationConfig, runConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the Ethereum nodeInfo wit service ID %v", serviceID)
	}

	logrus.Infof("Added Ethereum nodeInfo %v service with host port bindings: %+v ", serviceID, hostPortBindings)

	if err := networkCtx.WaitForHttpPostEndpointAvailability(serviceID, uint32(rpcPort), "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for service with ID '%v' to start", serviceID)
	}

	for _, nodeInfo := range nodesInfo {
		ok, err := addPeer(serviceCtx.GetIPAddress(), nodeInfo.Enode)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to call addPeer endpoint to add peer with nodeInfo %v", nodeInfo)
		}
		if !ok {
			return nil, stacktrace.NewError("addPeer endpoint returned false on service %v, adding peer %v", serviceID, nodeInfo)
		}
	}

	cmd := "geth attach data/geth.ipc --exec admin.peers"
	exitCode, logOutput, err := serviceCtx.ExecCommand([]string{
		"/bin/sh",
		"-c",
		cmd,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Executing command '%v' returned an error", cmd)
	}
	if exitCode != execCommandSuccessExitCode {
		return nil, stacktrace.NewError("Executing command returned an failing exit code with logs: %+v", string(*logOutput))
	}

	if err = validatePeersQuantity(string(*logOutput), serviceID, nodesInfo); err != nil {
		return nil, stacktrace.Propagate(err, "Validate peers error")
	}

	enode, err := getEnodeAddress(serviceCtx.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get nodeInfo from peer %v", serviceID)
	}

	nodeInfo := &NodeInfo{
		ServiceID: serviceCtx.GetServiceID(),
		Enode:     enode,
	}

	return nodeInfo, nil
}

func getContainerCreationConfig() *services.ContainerCreationConfig {

	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		ethereumDockerImageName,
	).WithUsedPorts(
		map[string]bool{
			fmt.Sprintf("%v/tcp", rpcPort):       true,
			fmt.Sprintf("%v/tcp", wsPort):        true,
			fmt.Sprintf("%v/tcp", discoveryPort): true,
			fmt.Sprintf("%v/udp", discoveryPort): true,
		},
	).WithStaticFiles(static_files_consts.StaticFiles).Build()

	return containerCreationConfig

}

func getEthBootnodeRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	runConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		genesisFilepath, found := staticFileFilepaths[static_files_consts.GenesisStaticFileID]
		if !found {
			return nil, stacktrace.NewError("No filepath found for key '%v'; this is a bug in Kurtosis!", static_files_consts.GenesisStaticFileID)
		}

		passwordFilepath, found := staticFileFilepaths[static_files_consts.PasswordStaticFileID]
		if !found {
			return nil, stacktrace.NewError("No filepath found for key '%v'; this is a bug in Kurtosis!", static_files_consts.PasswordStaticFileID)
		}

		signerKeystoreFilepath, found := staticFileFilepaths[static_files_consts.SignerKeystoreFileID]
		if !found {
			return nil, stacktrace.NewError("No filepath found for key '%v'; this is a bug in Kurtosis!", static_files_consts.SignerKeystoreFileID)
		}

		keystoreFolder := filepath.Dir(signerKeystoreFilepath)

		ipNet := getIPNet(ipAddr)

		entryPointArgs := []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("geth init --datadir data %v && geth --keystore %v --datadir data --networkid 15 -http --http.api admin,eth,net,web3,miner,personal,txpool,debug --http.addr %v --http.corsdomain '*' --nat extip:%v --port %v --unlock 0x14f6136b48b74b147926c9f24323d16c1e54a026 --mine --allow-insecure-unlock --netrestrict %v --password %v", genesisFilepath, keystoreFolder, ipAddr, ipAddr, discoveryPort, ipNet, passwordFilepath),
		}

		result := services.NewContainerRunConfigBuilder().WithEntrypointOverride(entryPointArgs).Build()
		return result, nil
	}
	return runConfigFunc

}

func getEthNodeRunConfigFunc(bootnodeEnr string) func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	runConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		genesisFilepath, found := staticFileFilepaths[static_files_consts.GenesisStaticFileID]
		if !found {
			return nil, stacktrace.NewError("No filepath found for test file 1 key '%v'; this is a bug in Kurtosis!", static_files_consts.GenesisStaticFileID)
		}

		entryPointArgs := []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("geth init --datadir data %v && geth --datadir data --networkid 15 -http --http.api admin,eth,net,web3,miner,personal,txpool,debug --http.addr %v --http.corsdomain '*' --nat extip:%v --gcmode archive --syncmode full --port %v --bootnodes %v", genesisFilepath, ipAddr, ipAddr, discoveryPort, bootnodeEnr),
		}

		result := services.NewContainerRunConfigBuilder().WithEntrypointOverride(entryPointArgs).Build()
		return result, nil
	}
	return runConfigFunc
}

func getIPNet(ipAddr string) *net.IPNet {
	cidr := ipAddr + subnetRange
	_, ipNet, _ := net.ParseCIDR(cidr)
	return ipNet
}

func sendRpcCall(ipAddress string, rpcJsonString string, targetStruct interface{}) error {
	url := fmt.Sprintf("http://%v:%v", ipAddress, rpcPort)
	var jsonByteArray = []byte(rpcJsonString)

	logrus.Debugf("Sending RPC call to '%v' with JSON body '%v'...", url, rpcJsonString)

	client := http.Client{
		Timeout: rpcRequestTimeout,
	}
	resp, err := client.Post(url, jsonContentType, bytes.NewBuffer(jsonByteArray))
	if err != nil {
		return stacktrace.Propagate(err, "Failed to send RPC request to Geth node with ip '%v'", ipAddress)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// For debugging
		var teeBuf bytes.Buffer
		tee := io.TeeReader(resp.Body, &teeBuf)
		bodyBytes, err := ioutil.ReadAll(tee)
		if err != nil {
			return stacktrace.Propagate(err, "Error parsing Geth node response into bytes.")
		}
		bodyString := string(bodyBytes)
		logrus.Tracef("Response for RPC call %v: %v", rpcJsonString, bodyString)

		err = json.NewDecoder(&teeBuf).Decode(targetStruct)
		if err != nil {
			return stacktrace.Propagate(err, "Error parsing geth node response into target struct.")
		}
		return nil
	} else {
		return stacktrace.NewError("Received non-200 status code rom admin RPC api: %v", resp.StatusCode)
	}
}

func addPeer(ipAddress string, peerEnode string) (bool, error) {
	adminAddPeerRpcCall := fmt.Sprintf(`{"jsonrpc":"2.0", "method": "admin_addPeer", "params": ["%v"], "id":70}`, peerEnode)
	logrus.Infof("Admin add peer rpc call: %v", adminAddPeerRpcCall)
	addPeerResponse := new(AddPeerResponse)
	err := sendRpcCall(ipAddress, adminAddPeerRpcCall, addPeerResponse)
	logrus.Infof("addPeer response: %+v", addPeerResponse)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to send addPeer RPC call for enode %v", peerEnode)
	}
	return addPeerResponse.Result, nil
}

func validatePeersQuantity(logString string, serviceID services.ServiceID, nodesInfo []*NodeInfo) error {
	peersQuantity := strings.Count(logString, enodePrefix) - strings.Count(logString, handshakeProtocol)
	validPeersQuantity := len(nodesInfo) + 1
	if peersQuantity != validPeersQuantity {
		return stacktrace.NewError("The amount of peers '%v' for node '%v' is not valid, should be '%v?", peersQuantity, serviceID, validPeersQuantity)
	}
	return nil
}

func getEnodeAddress(ipAddress string) (string, error) {
	nodeInfoResponse := new(NodeInfoResponse)
	err := sendRpcCall(ipAddress, adminInfoRpcCall, nodeInfoResponse)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send admin node info RPC request to Geth node with ip %v", ipAddress)
	}
	return nodeInfoResponse.Result.Enode, nil
}
