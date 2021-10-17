# TBD

# 0.3.0
### Features
* Run the module in CI, for extra verification

### Changes
* Made the instructions in the README for running the module simpler
* Use the new module API lib, which replaces all references of "Lambda" with "module"

### Removals
* Removed the world-public download token in CircleCI config when installing Kurtosis CLI, as it's no longer needed

### Breaking Changes
* Upgrade to the [module API lib 0.10.0](https://github.com/kurtosis-tech/kurtosis-module-api-lib/blob/develop/docs/changelog.md#0100), which replaces all references of "Lambda" with "module"
    * Users will need the latest version of Kurtosis CLI which has `module exec` rather than `lambda exec` to run this NEAR module

# 0.2.5
### Features
* Upgraded to [Lambda API Lib 0.9.2](https://github.com/kurtosis-tech/kurtosis-lambda-api-lib/blob/develop/docs/changelog.md#092)

# 0.2.4
### Features
* Upgraded Kurt Lamba API Lib dependency to the latest version [Kurt Lambda API Lib 0.9.1](https://github.com/kurtosis-tech/kurtosis-lambda-api-lib/blob/develop/docs/changelog.md#091)

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

