package impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type EthereumKurtosisLambda struct {
}

type EthereumKurtosisLambdaParams struct {
	IWantATip bool `json:"iWantATip"`
}

type EthereumKurtosisLambdaResult struct {
	Tip string `json:"tip"`
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

	ethereumKurtosisLambdaResult := &EthereumKurtosisLambdaResult{
	}

	result, err := json.Marshal(ethereumKurtosisLambdaResult)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the Ethereum Kurtosis Lambda Result with value '%+v'", ethereumKurtosisLambdaResult)
	}
	stringResult := string(result)

	logrus.Info("Ethereum Kurtosis Lambda executed successfully")
	return stringResult, nil
}
