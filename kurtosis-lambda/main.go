package main

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang/lib/execution"
	"github.com/kurtosis-tech/ethereum-kurtosis-lambda/kurtosis-lambda/impl"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {

	// >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<
	configurator := impl.NewExampleKurtosisLambdaConfigurator()
	// >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<


	lambdaExecutor := execution.NewKurtosisLambdaExecutor(configurator)
	if err := lambdaExecutor.Run(); err != nil {
		logrus.Errorf("An error occurred running the Kurtosis Lambda executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
