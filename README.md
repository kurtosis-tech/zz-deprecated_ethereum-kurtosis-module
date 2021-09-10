Ethereum Kurtosis Lambda
=====================
You can use the Ethereum Kurtosis Lambda to run and interact with an Ethereum private network inside a Kurtosis testsuite or by Kurtosis interactive.

###How to load, start and interact with Ethereum Kurtosis Lambda in a Kurtosis testsuite written with Golang

1. Load an Ethereum Kurtosis Lambda in a Kurtosis testsuite
   1. Add the following sentence `networkCtx.LoadLambda("ethereum-kurtosis-lambda","kurtosistech/ethereum-kurtosis-lambda","{}")` in the `Setup` method of the Kurtosis test
1. Start an Ethereum private network in a Kurtosis testsuite
   1. Get the Ethereum Kurtosis Lambda context to get lambda's context information
      1. First, cast the network object received in the `Run` method of the Kurtosis test to a network context object
         1. Add the following sentence `networkCtx := network.(*networks.NetworkContext)` in the `Run` method
      1. Get the lambda context
         1. Add the following sentence `ethLambdaCtx, err := networkCtx.GetLambdaContext("ethereum-kurtosis-lambda")`
         1. Do not forget to control the `err` var
      1. Start an Ethereum private network which contains a bootnode and other three nodes
         1. Add the following sentence `respJsonStr, err := ethLambdaCtx.Execute("{}")` to start the network and receive a json string object with Ethereum network context information with the following structure:
         ``
            {
               "bootnode_service_id":"bootnode",
               "node_service_ids":[
                  "ethereum-node-1",
                  "ethereum-node-2",
                  "ethereum-node-3"
               ],
               "genesis_static_file_id":"genesis",
               "password_static_file_id":"password",
               "signer_keystore_static_file_id":"signer-key"
            }
         ``
1. Interact with an Ethereum Kurtosis Lambda in a Kurtosis testsuite
   1. Cast the json string value returned by the `Execute` method to a struct
      1. Add the `EthereumKurtosisLambdaResult` struct
      ``
         type EthereumKurtosisLambdaResult struct {
            BootnodeServiceID          services.ServiceID      `json:"bootnode_service_id"`
            NodeServiceIDs             []services.ServiceID    `json:"node_service_ids"`
            StaticFileIDs              []services.StaticFileID `json:"static_file_ids"`
            GenesisStaticFileID        services.StaticFileID   `json:"genesis_static_file_id"`
            PasswordStaticFileID       services.StaticFileID   `json:"password_static_file_id"`
            SignerKeystoreStaticFileID services.StaticFileID   `json:"signer_keystore_static_file_id"`
         }
      ``
      1. Unmarshall the json string value to the `EthereumKurtosisLambdaResult`
      ``
         ethResult := new(EthereumKurtosisLambdaResult)
         if err := json.Unmarshal([]byte(respJsonStr), ethResult); err != nil {
            return stacktrace.Propagate(err, "An error occurred deserializing the Lambda response")
         }
      ``
   1. Get the Ethereum bootnode service context
      1. Add the following sentence `bootnodeServiceCtx, err := castedNetwork.GetServiceContext(ethResult.BootnodeServiceID)` to get the Ethereum bootnode context
      1. Do not forget to control the `err` var
   1. Execute a Geth command inside the Ethereum bootnode service
      1. Execute the Geth `eth.accounts` command to list the Ethereum network accounts
      ``
         gethCmd := "geth attach data/geth.ipc --exec eth.accounts"
         listAccountsCmd := []string{
            "/bin/sh",
            "-c",
            gethCmd,
         }

         exitCode, logOutput, err := serviceCtx.ExecCommand(listAccountsCmd)
         if err != nil {
            return stacktrace.Propagate(err, "Executing command '%v' returned an error", gethCmd)
         }
         if exitCode != execCommandSuccessExitCode {
            return stacktrace.NewError("Executing command returned an failing exit code with logs: %+v", string(*logOutput))
         }
      ``
      1. Read the [official Geth documentation](https://geth.ethereum.org/docs/) to discover and learn what others commands are available to execute in an Ethereum node