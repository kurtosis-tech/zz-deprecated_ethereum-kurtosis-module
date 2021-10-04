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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ethereumDockerImageName = "ethereum/client-go:v1.10.8"

	rpcPort       uint32 = 8545
	wsPort        uint32 = 8546
	discoveryPort uint32 = 30303
	subnetRange          = "/24"

	bootnodeServiceID           = "bootnode"
	childEthNodeServiceIdPrefix = "ethereum-node-"
	childEthNodeQuantity        = 2

	waitEndpointInitialDelayMilliseconds = 100
	waitEndpointRetries                  = 30
	waitEndpointRetriesDelayMilliseconds = 500

	adminInfoRpcCall           = `{"jsonrpc":"2.0","method": "admin_nodeInfo","params":[],"id":67}`
	execCommandSuccessExitCode = 0
	rpcRequestTimeout          = 30 * time.Second
	jsonContentType            = "application/json"
	enodePrefix                = "enode://"
	handshakeProtocol          = "eth: \"handshake\""

	ethNetworkId = 77813

	maxNumPeerCountValidationAttempts      = 5
	timeBetweenPeerCountValidationAttempts = 500 * time.Millisecond
)

var usedPortsSet = map[string]bool{
	fmt.Sprintf("%v/tcp", rpcPort):               true,
	fmt.Sprintf("%v/tcp", wsPort):        true,
	fmt.Sprintf("%v/tcp", discoveryPort): true,
	fmt.Sprintf("%v/udp", discoveryPort): true,
}

type EthereumKurtosisLambda struct {
}

func NewEthereumKurtosisLambda() *EthereumKurtosisLambda {
	return &EthereumKurtosisLambda{}
}

func (e EthereumKurtosisLambda) Execute(networkCtx *networks.NetworkContext, serializedParams string) (serializedResult string, resultError error) {
	logrus.Infof("Ethereum Kurtosis Lambda receives serializedParams '%v'", serializedParams)
	serializedParamsBytes := []byte(serializedParams)
	var params LambdaAPIExecuteArgs
	if err := json.Unmarshal(serializedParamsBytes, &params); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the Ethereum Kurtosis Lambda serialized params with value '%v'", serializedParams)
	}

	allNodeInfo, bootnodeServiceCtx, err := startEthNodes(networkCtx)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the Ethereum child nodes")
	}

	signerKeystoreContent, err := getStaticFileContent(bootnodeServiceCtx, static_files_consts.SignerKeystoreFileName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting an static file content '%v'", static_files_consts.SignerKeystoreFileName)
	}

	signerAccountPasswordContent, err := getStaticFileContent(bootnodeServiceCtx, static_files_consts.SignerAccountPasswordStaticFileName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting an static file content '%v'", static_files_consts.SignerAccountPasswordStaticFileName)
	}

	ethereumKurtosisLambdaResult := &LambdaAPIExecuteResult{
		BootnodeServiceID:     bootnodeServiceID,
		NodeInfo:              allNodeInfo,
		SignerKeystoreContent: signerKeystoreContent,
		SignerAccountPassword: signerAccountPasswordContent,
	}

	result, err := json.Marshal(ethereumKurtosisLambdaResult)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the Ethereum Kurtosis Lambda Result with value '%+v'", ethereumKurtosisLambdaResult)
	}
	stringResult := string(result)

	logrus.Infof("Ethereum Kurtosis Lambda Result value: %+v", ethereumKurtosisLambdaResult)
	logrus.Info("Ethereum Kurtosis Lambda executed successfully")
	return stringResult, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func startEthBootnode(networkCtx *networks.NetworkContext) (
	nodeServiceCtx *services.ServiceContext,
	enr string,
	nodeInfo *LambdaAPINodeInfo,
	resultErr error,
) {
	containerConfigSupplier := getBootnodeContainerConfigSupplier()

	serviceCtx, hostPortBindings, err := networkCtx.AddService(bootnodeServiceID, containerConfigSupplier)
	if err != nil {
		return nil, "", nil, stacktrace.Propagate(err, "An error occurred adding the Ethereum bootnode service")
	}

	if err := networkCtx.WaitForHttpPostEndpointAvailability(bootnodeServiceID, rpcPort, "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {
		return nil, "", nil, stacktrace.Propagate(err, "An error occurred waiting for service with ID '%v' to start", bootnodeServiceID)
	}

	logrus.Infof("Added Ethereum bootnode service with IP: %v and host port bindings: %+v", serviceCtx.GetIPAddress(), hostPortBindings)

	cmd := "geth attach data/geth.ipc --exec admin.nodeInfo.enr"
	exitCode, logOutput, err := serviceCtx.ExecCommand([]string{
		"/bin/sh",
		"-c",
		cmd,
	})
	if err != nil {
		return nil, "", nil, stacktrace.Propagate(err, "Executing command '%v' returned an error", cmd)
	}
	if exitCode != execCommandSuccessExitCode {
		return nil, "", nil, stacktrace.NewError("Executing command '%v' returned an failing exit code with logs:\n%v", cmd, logOutput)
	}

	lambdaApiNodeInfo := &LambdaAPINodeInfo{
		IPAddrInsideNetwork:        serviceCtx.GetIPAddress(),
		ExposedPortsSet:            usedPortsSet,
		PortBindingsOnLocalMachine: hostPortBindings,
	}

	return serviceCtx, logOutput, lambdaApiNodeInfo, nil
}

func startEthNodes(
	networkCtx *networks.NetworkContext,
) (map[services.ServiceID]*LambdaAPINodeInfo, *services.ServiceContext, error) {
	bootnodeServiceCtx, bootnodeEnr, bootnodeInfo, err := startEthBootnode(networkCtx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the Ethereum bootnode")
	}

	// Start all child nodes without waiting for them to become available, to speed up startup
	childNodeInfo := map[services.ServiceID]*LambdaAPINodeInfo{}
	allNodeServiceCtxs := map[services.ServiceID]*services.ServiceContext{
		bootnodeServiceID: bootnodeServiceCtx,
	}
	for i := 1; i <= childEthNodeQuantity; i++ {
		serviceId := services.ServiceID(childEthNodeServiceIdPrefix + strconv.Itoa(i))

		containerConfigSupplier := getEthNodeContainerConfigSupplier(bootnodeEnr)

		serviceCtx, hostPortBindings, err := networkCtx.AddService(serviceId, containerConfigSupplier)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred adding Ethereum node with service ID '%v'", serviceId)
		}
		logrus.Infof("Added Ethereum node '%v' with host port bindings: %+v ", serviceId, hostPortBindings)


		lambdaApiNodeInfo := &LambdaAPINodeInfo{
			IPAddrInsideNetwork:        serviceCtx.GetIPAddress(),
			ExposedPortsSet:            usedPortsSet,
			PortBindingsOnLocalMachine: hostPortBindings,
		}

		childNodeInfo[serviceId] = lambdaApiNodeInfo
		allNodeServiceCtxs[serviceId] = serviceCtx
	}

	// Now after all child nodes are started, wait for them to become available
	for childServiceId := range childNodeInfo {
		if err := networkCtx.WaitForHttpPostEndpointAvailability(childServiceId, rpcPort, "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for child node with ID '%v' to start", childServiceId)
		}
	}

	// Get the child ENRs, for use in adding peers...
	childEnodeAddrs := map[services.ServiceID]string{}
	peersToConnectPerNode := map[services.ServiceID][]string{}
	for childServiceId := range childNodeInfo {
		childServiceCtx, found := allNodeServiceCtxs[childServiceId]
		if !found {
			return nil, nil, stacktrace.NewError("No service context found for child node '%v'; this is a bug with this Lambda", childServiceId)
		}

		childPeers := []string{}
		for _, peerEnode := range childEnodeAddrs {
			childPeers = append(childPeers, peerEnode)
		}
		peersToConnectPerNode[childServiceId] = childPeers

		enodeAddr, err := getEnodeAddress(childServiceCtx.GetIPAddress())
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't get enode address for child node '%v'", childServiceId)
		}
		childEnodeAddrs[childServiceId] = enodeAddr
	}

	// ...and connect all the peers together, because Geth gossip is sloww
	for childServiceId, childPeersToConnect := range peersToConnectPerNode {
		childServiceCtx, found := allNodeServiceCtxs[childServiceId]
		if !found {
			return nil, nil, stacktrace.NewError("No service context for child '%v'; this is a bug with this Lambda", childServiceId)
		}
		for _, peerEnode := range childPeersToConnect {
			if err := addPeer(childServiceCtx.GetIPAddress(), peerEnode); err != nil {
				return nil, nil, stacktrace.Propagate(
					err,
					"An error occurred connecting peer enode '%v' to child with service ID '%v'",
					peerEnode,
					childServiceId,
				)
			}
		}
	}

	// Finally, verify that each node has the correct number of peers
	numExpectedPeersPerNode := len(allNodeServiceCtxs) - 1
	for serviceId, serviceCtx := range allNodeServiceCtxs {
		isPeerCountValidated := false
		for i := 0; i < maxNumPeerCountValidationAttempts; i++ {
			if verifyErr := verifyExpectedNumberPeers(serviceId, serviceCtx, numExpectedPeersPerNode); verifyErr == nil {
				isPeerCountValidated = true
				break
			} else {
				logrus.Debugf(
					"Verifying expected number of peers on node '%v' failed with error:\n%v",
					serviceId,
					verifyErr,
				)
				time.Sleep(timeBetweenPeerCountValidationAttempts)
			}
		}
		if !isPeerCountValidated {
			return nil, nil, stacktrace.NewError(
				"Service '%v' didn't reach expected number of peers '%v', even after %v attempts with %v between attempts",
				serviceId,
				numExpectedPeersPerNode,
				maxNumPeerCountValidationAttempts,
				timeBetweenPeerCountValidationAttempts,
			)
		}
	}

	allNodeInfo := map[services.ServiceID]*LambdaAPINodeInfo{
		bootnodeServiceID: bootnodeInfo,
	}
	for childServiceId, childInfo := range childNodeInfo {
		allNodeInfo[childServiceId] = childInfo
	}

	return allNodeInfo, bootnodeServiceCtx, nil
}

func verifyExpectedNumberPeers(serviceId services.ServiceID, serviceCtx *services.ServiceContext, numExpectedPeers int) error {
	cmd := "geth attach data/geth.ipc --exec admin.peers"
	exitCode, logOutput, err := serviceCtx.ExecCommand([]string{
		"/bin/sh",
		"-c",
		cmd,
	})
	if err != nil {
		return stacktrace.Propagate(
			err,
			"Executing peer-getting command '%v' on service '%v' returned an error",
			cmd,
			serviceId,
		)
	}
	if exitCode != execCommandSuccessExitCode {
		return stacktrace.NewError(
			"Executing peer-getting command '%v' on service '%v' returned non-%v exit code '%v' with the following logs:\n%v",
			cmd,
			serviceId,
			execCommandSuccessExitCode,
			exitCode,
			logOutput,
		)
	}

	// peersQuantity := strings.Count(logOutputStr, enodePrefix) - strings.Count(logOutputStr, handshakeProtocol)
	peersQuantity := strings.Count(logOutput, enodePrefix) - strings.Count(logOutput, handshakeProtocol)
	if peersQuantity != numExpectedPeers {
		return stacktrace.NewError(
			"Expected '%v' peers for node '%v' but got '%v'",
			numExpectedPeers,
			serviceId,
			peersQuantity,
		)
	}
	return nil
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

// Geth gossiping is slowww, so we manually add nodes to speed it up
func addPeer(ipAddress string, peerEnode string) error {
	adminAddPeerRpcCall := fmt.Sprintf(`{"jsonrpc":"2.0", "method": "admin_addPeer", "params": ["%v"], "id":70}`, peerEnode)
	logrus.Infof("Admin add peer rpc call: %v", adminAddPeerRpcCall)
	addPeerResponse := new(EthAPIAddPeerResponse)
	err := sendRpcCall(ipAddress, adminAddPeerRpcCall, addPeerResponse)
	logrus.Infof("addPeer response: %+v", addPeerResponse)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to send addPeer RPC call for enode %v", peerEnode)
	}
	if addPeerResponse.Result == false {
		return stacktrace.NewError(
			"Ethereum returned 'false' response to addPeer request to add enode '%v' to node with IP '%v'",
			peerEnode,
			ipAddress,
		)
	}
	return nil
}

func getEnodeAddress(ipAddress string) (string, error) {
	nodeInfoResponse := new(EthAPINodeInfoResponse)
	err := sendRpcCall(ipAddress, adminInfoRpcCall, nodeInfoResponse)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send admin node info RPC request to Geth node with ip %v", ipAddress)
	}
	return nodeInfoResponse.Result.Enode, nil
}

func getStaticFileContent(serviceCtx *services.ServiceContext, staticFileName string) (string, error) {

	staticFileFilePath := serviceCtx.GetSharedDirectory().GetChildPath(staticFileName)

	catStaticFileCmd := []string{
		"cat",
		staticFileFilePath.GetAbsPathOnServiceContainer(),
	}
	exitCode, fileContents, err := serviceCtx.ExecCommand(catStaticFileCmd)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static '%v' contents", catStaticFileCmd, staticFileName)
	}
	if exitCode != execCommandSuccessExitCode {
		return "", stacktrace.NewError("Command '%+v' to cat the static file '%v' exited with non-successful exit code '%v'", catStaticFileCmd, staticFileFilePath.GetAbsPathOnServiceContainer(), exitCode)
	}

	return fileContents, nil
}

func getBootnodeContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		//Copy static files from the static_files folder in testsuite container to the service's folder in the service container
		if err := copyStaticFilesInServiceContainer(static_files_consts.StaticFilesNames, static_files_consts.StaticFilesDirpathOnTestsuiteContainer, sharedDirectory); err != nil{
			return nil, stacktrace.Propagate(err, "An error occurred copying static files into the service's folder in the service container")
		}

		keystoreFolder := filepath.Dir(sharedDirectory.GetChildPath(static_files_consts.SignerKeystoreFileName).GetAbsPathOnServiceContainer())

		ipNet := getIPNet(ipAddr)

		entryPointArgs := []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf(
				"geth init --datadir data %v && " +
					"geth " +
					"--keystore %v " +
					"--datadir data " +
					"--networkid %v " +
					"-http " +
					"--http.api admin,eth,net,web3,miner,personal,txpool,debug " +
					"--http.addr %v " +
					"--http.corsdomain '*' " +
					"--nat extip:%v " +
					"--port %v " +
					"--unlock 0x14f6136b48b74b147926c9f24323d16c1e54a026 --" +
					"mine " +
					"--allow-insecure-unlock " +
					"--netrestrict %v " +
					"--password %v",
				sharedDirectory.GetChildPath(static_files_consts.GenesisStaticFileName).GetAbsPathOnServiceContainer(),
				keystoreFolder,
				ethNetworkId,
				ipAddr,
				ipAddr,
				discoveryPort,
				ipNet,
				sharedDirectory.GetChildPath(static_files_consts.SignerAccountPasswordStaticFileName).GetAbsPathOnServiceContainer(),
			),
		}

		containerConfig := services.NewContainerConfigBuilder(
			ethereumDockerImageName,
		).WithUsedPorts(
			usedPortsSet,
		).WithEntrypointOverride(
			entryPointArgs,
	    ).Build()

		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getEthNodeContainerConfigSupplier(bootnodeEnr string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		//Copy static files from the static_files folder in testsuite container to the service's folder in the service container
		staticFileNames := []string{static_files_consts.GenesisStaticFileName}
		if err := copyStaticFilesInServiceContainer(staticFileNames, static_files_consts.StaticFilesDirpathOnTestsuiteContainer, sharedDirectory); err != nil{
			return nil, stacktrace.Propagate(err, "An error occurred copying static files into the service's folder in the service container")
		}

		entryPointArgs := []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf(
				"geth init --datadir data %v && " +
					"geth " +
					"--datadir data " +
					"--networkid %v " +
					"-http " +
					"--http.api admin,eth,net,web3,miner,personal,txpool,debug " +
					"--http.addr %v " +
					"--http.corsdomain '*' " +
					"--nat extip:%v " +
					"--gcmode archive " +
					"--syncmode full " +
					"--port %v " +
					"--bootnodes %v",
				sharedDirectory.GetChildPath(static_files_consts.GenesisStaticFileName).GetAbsPathOnServiceContainer(),
				ethNetworkId,
				ipAddr,
				ipAddr,
				discoveryPort,
				bootnodeEnr,
			),
		}

		containerConfig := services.NewContainerConfigBuilder(
			ethereumDockerImageName,
		).WithUsedPorts(
			usedPortsSet,
		).WithEntrypointOverride(
			entryPointArgs,
		).Build()

		return containerConfig, nil
	}
	return containerConfigSupplier
}

func copyStaticFilesInServiceContainer(staticFilesNames []string, staticFilesFolder string, sharedDirectory *services.SharedPath) error {
	for _, staticFileName := range staticFilesNames {
		if err := copyStaticFileInServiceContainer(staticFileName, staticFilesFolder, sharedDirectory); err != nil {
			return stacktrace.Propagate(err, "An error occurred copying file with filename '%v' to service's folder in service container", staticFileName)
		}
	}
	return nil
}

func copyStaticFileInServiceContainer(staticFileName string, staticFilesFolder string,sharedDirectory *services.SharedPath) error {
	testStaticFileFilePath := sharedDirectory.GetChildPath(staticFileName)

	testStaticFilepath := filepath.Join(staticFilesFolder, staticFileName)

	testStaticFileContent, err := ioutil.ReadFile(testStaticFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading file from '%v'", testStaticFilepath)
	}

	err = ioutil.WriteFile(testStaticFileFilePath.GetAbsPathOnThisContainer(), testStaticFileContent, os.ModePerm)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing file '%v'", testStaticFileFilePath.GetAbsPathOnThisContainer())
	}
	return nil
}
