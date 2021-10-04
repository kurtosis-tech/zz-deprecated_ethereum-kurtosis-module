package static_files_consts

const (
	// Directory where static files live inside the testsuite container
	StaticFilesDirpathOnTestsuiteContainer = "/static-files"

	GenesisStaticFileName               = "genesis.json"
	SignerAccountPasswordStaticFileName = "password.txt"
	SignerKeystoreFileName              = "UTC--2021-08-11T21-30-29.861585000Z--14f6136b48b74b147926c9f24323d16c1e54a026"
)

var StaticFilesNames = []string{GenesisStaticFileName, SignerAccountPasswordStaticFileName, SignerKeystoreFileName}
