package impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

var (
	tipsRepository = []string{
		"Everything not saved will be lost.",
		"Don't pet a burning dog.",
		"Even a broken clock is right twice a day.",
		"If no one comes from the future to stop you from doing it, then how bad of a decision can it really be?",
		"Never fall in love with a tennis player. Love means nothing to them.",
		"If you ever get caught sleeping on the job, slowly raise your head and say 'In Jesus' name, Amen'",
		"Never trust in an electrician with no eyebrows",
		"If you sleep until lunch time, you can save the breakfast money.",
	}
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
		Tip: getRandomTip(params.IWantATip),
	}

	result, err := json.Marshal(ethereumKurtosisLambdaResult)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the Ethereum Kurtosis Lambda Result with value '%+v'", ethereumKurtosisLambdaResult)
	}
	stringResult := string(result)

	logrus.Info("Ethereum Kurtosis Lambda executed successfully")
	return stringResult, nil
}

func getRandomTip(shouldGiveAdvice bool) string {
	var tip string
	if shouldGiveAdvice {
		rand.Seed(time.Now().Unix())
		tip = tipsRepository[rand.Intn(len(tipsRepository))]
	} else {
		tip = "Kurtosis Lambda won't enlighten you today."
	}
	return tip
}
