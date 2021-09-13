package static_files_consts

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"path"
)

const (
	// Directory where static files live inside the testsuite container
	staticFilesDirpathOnTestsuiteContainer = "/static-files"

	GenesisStaticFileID               services.StaticFileID = "genesis"
	SignerAccountPasswordStaticFileID services.StaticFileID = "password"
	SignerKeystoreFileID              services.StaticFileID = "signer-key"

	genesisStaticFileName               = "genesis.json"
	signerAccountPasswordStaticFileName = "password.txt"
	signerKeystoreFileName              = "UTC--2021-08-11T21-30-29.861585000Z--14f6136b48b74b147926c9f24323d16c1e54a026"
)

var StaticFiles = map[services.StaticFileID]bool{
	GenesisStaticFileID:               true,
	SignerAccountPasswordStaticFileID: true,
	SignerKeystoreFileID:              true,
}

var StaticFileFilepaths = map[services.StaticFileID]string{
	GenesisStaticFileID:               path.Join(staticFilesDirpathOnTestsuiteContainer, genesisStaticFileName),
	SignerAccountPasswordStaticFileID: path.Join(staticFilesDirpathOnTestsuiteContainer, signerAccountPasswordStaticFileName),
	SignerKeystoreFileID:              path.Join(staticFilesDirpathOnTestsuiteContainer, signerKeystoreFileName),
}
