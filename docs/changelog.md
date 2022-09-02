# TBD
### Changes
* Upgrade to core to 1.58.1 and module-api-lib to 0.21.1

# 0.5.6
### Changes
* Upgraded Ubuntu machine image in Circle CI configuration to version `ubuntu-2004:202201-02`
* Migrate repo to use internal cli tool, `kudet`, for release workflow and getting docker image tags
* Upgrade core to 1.55.2
* Upgrade module-api-lib to 0.17.0

# 0.5.5
### Changes
* Upgrades to module-api-lib 0.16.0 and core 1.54.1 which replace FilesArtifactID type with FilesArtifactUUID

# 0.5.4
### Changes
* Use module-api-lib 0.15.0

# 0.5.3
### Changes
* Use module-api-lib 0.14.1

# 0.5.2
### Changes
* Replaced the private IP address in the '--http.addr' with the 0.0.0.0 value in order to get access from local host
* Specified the '--http.port' default value 

# 0.5.1
### Fixes
* Bugfix so that the module works with Kurtosis CLI 0.11

# 0.5.0
### Features
* Return port IDs of the service ports so that users can get public & private ports by themselves

### Breaking Changes
* Removed the following fields from the response object:
    * `rpcPortOnHostMachine` -> `rpc_port_on_host_machine`
    * `wsPortOnHostMachine` -> `ws_port_on_host_machine`
    * `tcpDiscoveryPortOnHostMachine` -> `tcp_discovery_port_on_host_machine`
    * `udpDiscoveryPortOnHostMachine` -> `udp_discovery_port_on_host_machine`
* Added the following fields to the response object:
    * `rpc_port_id`
    * `ws_port_id`
    * `tcp_discovery_port_id`
    * `udp_discovery_port_id`

# 0.4.0
### Features
* Upgrade to module-api-lib 0.12.3, to support the latest version of Kurtosis
* Output JSON is now pretty-printed

### Breaking Changes
* The JSON object reported by the module to represent each node used to have `ExposedPortsSet` and `PortBindingsOnLocalMachine` maps, but now has:
    * `IPAddrOnHostMachine`
    * `RpcPortOnHostMachine`
    * `WsPortOnHostMachine`
    * `TcpDiscoveryPortOnHostMachine`
    * `UdpDiscoveryPortOnHostMachine`

# 0.3.2
### Fixes
* Upgrade to the latest Module API lib, as the module is currently broken

# 0.3.1
### Fixes
* Fix broken README links

# 0.3.0
### Features
* Run the module in CI, for extra verification

### Changes
* Made the instructions in the README for running the module simpler
* Use the new module API lib, which replaces all references of "Lambda" with "module"

### Removals
* Removed the world-public download token in CircleCI config when installing Kurtosis CLI, as it's no longer needed

### Breaking Changes
* Upgrade to the module API lib 0.10.0 which replaces all references of "Lambda" with "module"
    * Users will need the latest version of Kurtosis CLI which has `module exec` rather than `lambda exec` to run this NEAR module

# 0.2.5
### Features
* Upgraded to Lambda API Lib 0.9.2

# 0.2.4
### Features
* Upgraded Kurt Lamba API Lib dependency to the latest version Kurt Lambda API Lib 0.9.1

# 0.2.3
### Features
* Also publish a `latest` tag for this image

# 0.2.2
### Changes
* Simplified the README to show the Lambda being loaded into a Kurtosis sandbox

### Fixes
* Fixed an issue where, when waiting for ETH nodes to come up, we were waiting for 5ms between retries (changed to 500ms)

# 0.2.1
### Fixes
* Upgrade to lambda-api-lib 0.9.0, for latest Kurt Core compatibility

# 0.2.0
### Changes
* Start only 2 child ETH nodes (for 3 total, with the boot node) rather than 4
* Start all child ETH nodes then wait for them all to become available, rather than doing "start first node, wait for it, start second node, wait for it..."
* Changed the result object to also contain local host machine port bindings, so users can easily access the cluster in interactvie/debug mode

### Breaking Changes
* The Lambda execution result object has been completely redesigned to provide significantly more information
    * Users should switch to using the new object's fields

# 0.1.4
### Fixes
* Fixed bug with publishing the Docker image

# 0.1.3
### Changes
* Make versioning `X.Y.Z` rather than `X.Y`

# 0.1.2
### Changes
* Removed unnecessary (for now) CircleCI steps

# 0.1.1
### Fixes
* Added newline at the end of README file
* Corrected syntax error in changelog header

# 0.1.0
### Feature
* Implement a Kurtosis Lambda for an Ethereum Private Network ready to be used in Kurtosis testsuite or with Kurtosis Interactive
    * Set an Ethereum bootnode with a custom genesis block
    * Seth three Ethereum nodes using the bootnode
    * Connect peers between them and validate connectivity
    * Uses the Clique consensus algorithm to authorize transactions
* Add instructions in `README.md` file
* Add `get-docker-images-tag.sh` script
* Add `release.sh` script
* Add Circle CI configuration

