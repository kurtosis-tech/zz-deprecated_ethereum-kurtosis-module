package impl

import (
	"bytes"
	"encoding/json"
	"fmt"
	static_files_consts "github.com/kurtosis-tech/ethereum-kurtosis-module/kurtosis-module/static-files-consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	ethereumDockerImageName = "ethereum/client-go:v1.10.8"

	rpcPortNum       uint16 = 8545
	wsPortNum        uint16 = 8546
	discoveryPortNum uint16 = 30303
	subnetRange             = "/24"

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

	ethNetworkId = 881239

	maxNumPeerCountValidationAttempts      = 5
	timeBetweenPeerCountValidationAttempts = 500 * time.Millisecond

	// Port IDs
	rpcPortId          = "rpc"
	wsPortId           = "ws"
	tcpDiscoveryPortId = "tcpDiscovery"
	udpDiscoveryPortId = "udpDiscovery"

	jsonOutputPrefixStr = ""
	jsonOutputIndentStr = "  "

	staticFilesMountpointOnNodes = "/files"

	privateIPAddrPlaceholder = "KURTOSIS_PRIVATE_IP_ADDR_PLACEHOLDER"
)

var usedPorts = map[string]*services.PortSpec{
	rpcPortId:          services.NewPortSpec(rpcPortNum, services.PortProtocol_TCP),
	wsPortId:           services.NewPortSpec(wsPortNum, services.PortProtocol_TCP),
	tcpDiscoveryPortId: services.NewPortSpec(discoveryPortNum, services.PortProtocol_TCP),
	udpDiscoveryPortId: services.NewPortSpec(discoveryPortNum, services.PortProtocol_UDP),
}

type EthereumKurtosisModule struct {
}

func NewEthereumKurtosisModule() *EthereumKurtosisModule {
	return &EthereumKurtosisModule{}
}

func (e EthereumKurtosisModule) Execute(enclaveCtx *enclaves.EnclaveContext, serializedParams string) (serializedResult string, resultError error) {
	logrus.Infof("Serialized execute params '%v'", serializedParams)
	serializedParamsBytes := []byte(serializedParams)
	var params ModuleAPIExecuteArgs
	if err := json.Unmarshal(serializedParamsBytes, &params); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the serialized params with value '%v'", serializedParams)
	}

	staticFilesArtifactUuid, err := enclaveCtx.UploadFiles(static_files_consts.StaticFilesDirpathOnTestsuiteContainer)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred uploading the static files")
	}

	allNodeInfo, bootnodeServiceCtx, err := startEthNodes(enclaveCtx, staticFilesArtifactUuid)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the Ethereum child nodes")
	}

	signerKeystoreContent, err := getStaticFileContent(bootnodeServiceCtx, static_files_consts.SignerKeystoreFileName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting static file content '%v'", static_files_consts.SignerKeystoreFileName)
	}

	signerAccountPasswordContent, err := getStaticFileContent(bootnodeServiceCtx, static_files_consts.SignerAccountPasswordStaticFileName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting static file content '%v'", static_files_consts.SignerAccountPasswordStaticFileName)
	}

	resultObj := &ModuleAPIExecuteResult{
		BootnodeServiceID:     bootnodeServiceID,
		NodeInfo:              allNodeInfo,
		SignerKeystoreContent: signerKeystoreContent,
		SignerAccountPassword: signerAccountPasswordContent,
	}

	resultBytes, err := json.MarshalIndent(resultObj, jsonOutputPrefixStr, jsonOutputIndentStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the result object '%+v'", resultObj)
	}
	resultStr := string(resultBytes)

	logrus.Infof("Result string: %v", resultStr)
	logrus.Info("Ethereum Kurtosis module executed successfully")
	return resultStr, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func startEthBootnode(
	enclaveCtx *enclaves.EnclaveContext,
	staticFilesArtifactUuid services.FilesArtifactUUID,
) (
	nodeServiceCtx *services.ServiceContext,
	enr string,
	nodeInfo *ModuleAPIEthereumNodeInfo,
	resultErr error,
) {
	containerConfigSupplier := getBootnodeContainerConfig(staticFilesArtifactUuid)

	serviceCtx, err := enclaveCtx.AddService(bootnodeServiceID, containerConfigSupplier)
	if err != nil {
		return nil, "", nil, stacktrace.Propagate(err, "An error occurred adding the Ethereum bootnode service")
	}

	if err := enclaveCtx.WaitForHttpPostEndpointAvailability(bootnodeServiceID, uint32(rpcPortNum), "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {

		return nil, "", nil, stacktrace.Propagate(err, "An error occurred waiting for service with ID '%v' to start", bootnodeServiceID)
	}

	logrus.Infof(
		"Added Ethereum bootnode service with public IP: %v and public ports: %+v",
		serviceCtx.GetMaybePublicIPAddress(),
		serviceCtx.GetPublicPorts(),
	)

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

	apiNodeInfo, err := getApiNodeObjFromNodeServiceCtx(serviceCtx)
	if err != nil {
		return nil, "", nil, stacktrace.Propagate(err, "An error occurred getting the node info API object from the boot node's service context")
	}

	return serviceCtx, logOutput, apiNodeInfo, nil
}

func startEthNodes(
	enclaveCtx *enclaves.EnclaveContext,
	staticFilesArtifactUuid services.FilesArtifactUUID,
) (map[services.ServiceID]*ModuleAPIEthereumNodeInfo, *services.ServiceContext, error) {
	bootnodeServiceCtx, bootnodeEnr, bootnodeInfo, err := startEthBootnode(enclaveCtx, staticFilesArtifactUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the Ethereum bootnode")
	}

	// Start all child nodes without waiting for them to become available, to speed up startup
	childNodeInfo := map[services.ServiceID]*ModuleAPIEthereumNodeInfo{}
	allNodeServiceCtxs := map[services.ServiceID]*services.ServiceContext{
		bootnodeServiceID: bootnodeServiceCtx,
	}
	for i := 1; i <= childEthNodeQuantity; i++ {
		serviceId := services.ServiceID(childEthNodeServiceIdPrefix + strconv.Itoa(i))

		containerConfig := getEthNodeContainerConfig(bootnodeEnr, staticFilesArtifactUuid)

		serviceCtx, err := enclaveCtx.AddService(serviceId, containerConfig)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred adding Ethereum node with service ID '%v'", serviceId)
		}
		logrus.Infof(
			"Added Ethereum child node service with public IP: %v and public ports: %+v",
			serviceCtx.GetMaybePublicIPAddress(),
			serviceCtx.GetPublicPorts(),
		)

		apiNodeInfo, err := getApiNodeObjFromNodeServiceCtx(serviceCtx)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the node info API object from the service context of child node '%v'", serviceId)
		}

		childNodeInfo[serviceId] = apiNodeInfo
		allNodeServiceCtxs[serviceId] = serviceCtx
	}

	// Now after all child nodes are started, wait for them to become available
	for childServiceId := range childNodeInfo {
		if err := enclaveCtx.WaitForHttpPostEndpointAvailability(childServiceId, uint32(rpcPortNum), "", adminInfoRpcCall, waitEndpointInitialDelayMilliseconds, waitEndpointRetries, waitEndpointRetriesDelayMilliseconds, ""); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for child node with ID '%v' to start", childServiceId)
		}
	}

	// Get the child ENRs, for use in adding peers...
	childEnodeAddrs := map[services.ServiceID]string{}
	peersToConnectPerNode := map[services.ServiceID][]string{}
	for childServiceId := range childNodeInfo {
		childServiceCtx, found := allNodeServiceCtxs[childServiceId]
		if !found {
			return nil, nil, stacktrace.NewError("No service context found for child node '%v'; this is a bug with this module", childServiceId)
		}

		childPeers := []string{}
		for _, peerEnode := range childEnodeAddrs {
			childPeers = append(childPeers, peerEnode)
		}
		peersToConnectPerNode[childServiceId] = childPeers

		enodeAddr, err := getEnodeAddress(childServiceCtx.GetPrivateIPAddress())
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't get enode address for child node '%v'", childServiceId)
		}
		childEnodeAddrs[childServiceId] = enodeAddr
	}

	// ...and connect all the peers together, because Geth gossip is sloww
	for childServiceId, childPeersToConnect := range peersToConnectPerNode {
		childServiceCtx, found := allNodeServiceCtxs[childServiceId]
		if !found {
			return nil, nil, stacktrace.NewError("No service context for child '%v'; this is a bug with this module", childServiceId)
		}
		for _, peerEnode := range childPeersToConnect {
			if err := addPeer(childServiceCtx.GetPrivateIPAddress(), peerEnode); err != nil {
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

	allNodeInfo := map[services.ServiceID]*ModuleAPIEthereumNodeInfo{
		bootnodeServiceID: bootnodeInfo,
	}
	for childServiceId, childInfo := range childNodeInfo {
		allNodeInfo[childServiceId] = childInfo
	}

	return allNodeInfo, bootnodeServiceCtx, nil
}

func getApiNodeObjFromNodeServiceCtx(serviceCtx *services.ServiceContext) (*ModuleAPIEthereumNodeInfo, error) {
	return &ModuleAPIEthereumNodeInfo{
		IPAddrInsideNetwork: serviceCtx.GetPrivateIPAddress(),
		IPAddrOnHostMachine: serviceCtx.GetMaybePublicIPAddress(),
		RpcPortId:           rpcPortId,
		WsPortId:            wsPortId,
		TcpDiscoveryPortId:  tcpDiscoveryPortId,
		UdpDiscoveryPortId:  udpDiscoveryPortId,
	}, nil
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

func sendRpcCall(privateIpAddr string, rpcJsonString string, targetStruct interface{}) error {
	url := fmt.Sprintf("http://%v:%v", privateIpAddr, rpcPortNum)
	var jsonByteArray = []byte(rpcJsonString)

	logrus.Debugf("Sending RPC call to '%v' with JSON body '%v'...", url, rpcJsonString)

	client := http.Client{
		Timeout: rpcRequestTimeout,
	}
	resp, err := client.Post(url, jsonContentType, bytes.NewBuffer(jsonByteArray))
	if err != nil {
		return stacktrace.Propagate(err, "Failed to send RPC request to Geth node with ip '%v'", privateIpAddr)
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
func addPeer(privateIpAddr string, peerEnode string) error {
	adminAddPeerRpcCall := fmt.Sprintf(`{"jsonrpc":"2.0", "method": "admin_addPeer", "params": ["%v"], "id":70}`, peerEnode)
	logrus.Infof("Admin add peer rpc call: %v", adminAddPeerRpcCall)
	addPeerResponse := new(EthAPIAddPeerResponse)
	err := sendRpcCall(privateIpAddr, adminAddPeerRpcCall, addPeerResponse)
	logrus.Infof("addPeer response: %+v", addPeerResponse)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to send addPeer RPC call for enode %v", peerEnode)
	}
	if addPeerResponse.Result == false {
		return stacktrace.NewError(
			"Ethereum returned 'false' response to addPeer request to add enode '%v' to node with IP '%v'",
			peerEnode,
			privateIpAddr,
		)
	}
	return nil
}

func getEnodeAddress(privateIpAddr string) (string, error) {
	nodeInfoResponse := new(EthAPINodeInfoResponse)
	err := sendRpcCall(privateIpAddr, adminInfoRpcCall, nodeInfoResponse)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to send admin node info RPC request to Geth node with ip %v", privateIpAddr)
	}
	return nodeInfoResponse.Result.Enode, nil
}

func getStaticFileContent(serviceCtx *services.ServiceContext, staticFileName string) (string, error) {
	absFilepathOnNode := getMountedPathOnNodeContainer(staticFileName)
	catStaticFileCmd := []string{
		"cat",
		absFilepathOnNode,
	}
	exitCode, fileContents, err := serviceCtx.ExecCommand(catStaticFileCmd)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static '%v' contents", catStaticFileCmd, staticFileName)
	}
	if exitCode != execCommandSuccessExitCode {
		return "", stacktrace.NewError("Command '%+v' to cat the static file '%v' exited with non-successful exit code '%v'", catStaticFileCmd, absFilepathOnNode, exitCode)
	}

	return fileContents, nil
}

func getBootnodeContainerConfig(staticFilesArtifactUuid services.FilesArtifactUUID) *services.ContainerConfig {

	/*
		//Copy static files from the static_files folder in testsuite container to the service's folder in the service container
		if err := copyStaticFilesInServiceContainer(static_files_consts.StaticFilesNames, static_files_consts.StaticFilesDirpathOnTestsuiteContainer, sharedDirectory); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred copying static files into the service's folder in the service container")
		}

		keystoreFolder := filepath.Dir(sharedDirectory.GetChildPath(static_files_consts.SignerKeystoreFileName).GetAbsPathOnServiceContainer())
	*/

	entryPointArgs := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf(
			"geth init --datadir data %v && "+
				"geth "+
				"--keystore %v "+
				"--datadir data "+
				"--networkid %v "+
				"-http "+
				"--http.api admin,eth,net,web3,miner,personal,txpool,debug "+
				"--http.addr=0.0.0.0 "+
				"--http.port=%v "+
				"--http.corsdomain '*' "+
				"--http.vhosts=* "+
				"--nat extip:"+privateIPAddrPlaceholder+" "+
				"--port=%v "+
				"--unlock 0x14f6136b48b74b147926c9f24323d16c1e54a026 --"+
				"mine "+
				"--allow-insecure-unlock "+
				"--netrestrict:"+privateIPAddrPlaceholder+"/24"+
				"--password %v",
			getMountedPathOnNodeContainer(static_files_consts.GenesisStaticFileName),
			getMountedPathOnNodeContainer(""), // The keystore arg expects a directory containing keys
			ethNetworkId,
			rpcPortNum,
			discoveryPortNum,
			getMountedPathOnNodeContainer(static_files_consts.SignerAccountPasswordStaticFileName),
		),
	}

	containerConfig := services.NewContainerConfigBuilder(
		ethereumDockerImageName,
	).WithUsedPorts(
		usedPorts,
	).WithEntrypointOverride(
		entryPointArgs,
	).WithFiles(map[services.FilesArtifactUUID]string{
		staticFilesArtifactUuid: staticFilesMountpointOnNodes,
	}).Build()

	return containerConfig
}

func getEthNodeContainerConfig(
	bootnodeEnr string,
	staticFilesArtifactUUid services.FilesArtifactUUID,
) *services.ContainerConfig {

	/*
		//Copy static files from the static_files folder in testsuite container to the service's folder in the service container
		staticFileNames := []string{static_files_consts.GenesisStaticFileName}
		if err := copyStaticFilesInServiceContainer(staticFileNames, static_files_consts.StaticFilesDirpathOnTestsuiteContainer, sharedDirectory); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred copying static files into the service's folder in the service container")
		}

	*/

	entryPointArgs := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf(
			"geth init --datadir data %v && "+
				"geth "+
				"--datadir data "+
				"--networkid %v "+
				"-http "+
				"--http.api admin,eth,net,web3,miner,personal,txpool,debug "+
				"--http.addr=0.0.0.0 "+
				"--http.port=%v "+
				"--http.corsdomain '*' "+
				"--http.vhosts=* "+
				"--nat extip:"+privateIPAddrPlaceholder+" "+
				"--gcmode archive "+
				"--syncmode full "+
				"--port=%v "+
				"--bootnodes %v",
			getMountedPathOnNodeContainer(static_files_consts.GenesisStaticFileName),
			ethNetworkId,
			rpcPortNum,
			discoveryPortNum,
			bootnodeEnr,
		),
	}

	containerConfig := services.NewContainerConfigBuilder(
		ethereumDockerImageName,
	).WithUsedPorts(
		usedPorts,
	).WithEntrypointOverride(
		entryPointArgs,
	).WithFiles(map[services.FilesArtifactUUID]string{
		staticFilesArtifactUUid: staticFilesMountpointOnNodes,
	}).WithPrivateIPAddrPlaceholder(
		privateIPAddrPlaceholder,
	).Build()

	return containerConfig
}

func getMountedPathOnNodeContainer(staticFilename string) string {
	return path.Join(
		staticFilesMountpointOnNodes,
		staticFilename,
	)
}
