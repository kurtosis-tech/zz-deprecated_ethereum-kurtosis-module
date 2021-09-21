# TBD
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

