# TBD
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

