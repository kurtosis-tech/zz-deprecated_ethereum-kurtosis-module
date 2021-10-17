package main

import (
	"fmt"
	"github.com/kurtosis-tech/ethereum-kurtosis-module/kurtosis-module/impl"
	"github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang/lib/execution"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {

	configurator := impl.NewEthereumKurtosisLambdaConfigurator()

	lambdaExecutor := execution.NewKurtosisLambdaExecutor(configurator)
	if err := lambdaExecutor.Run(); err != nil {
		logrus.Errorf("An error occurred running the Kurtosis Lambda executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
