package impl

import (
	"encoding/json"
	kurtosis_lambda "github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang/lib/kurtosis-lambda"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const(
	defaultLogLevel = "info"
)

type ExampleKurtosisLambdaConfigurator struct{}

func NewExampleKurtosisLambdaConfigurator() *ExampleKurtosisLambdaConfigurator {
	return &ExampleKurtosisLambdaConfigurator{}
}

func (t ExampleKurtosisLambdaConfigurator) ParseParamsAndCreateKurtosisLambda(serializedCustomParamsStr string) (kurtosis_lambda.KurtosisLambda, error) {
	serializedCustomParamsBytes := []byte(serializedCustomParamsStr)
	var args ExampleKurtosisLambdaArgs
	if err := json.Unmarshal(serializedCustomParamsBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the Kurtosis Lambda serialized custom params with value '%v", serializedCustomParamsStr)
	}

	err := setLogLevel(args.LogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred setting the log level")
	}

	lambda := NewExampleKurtosisLambda()

	return lambda, nil
}

func setLogLevel(logLevelStr string) error {
	if logLevelStr == "" {
		logLevelStr = defaultLogLevel
	}
	level, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing loglevel string '%v'", logLevelStr)
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return nil
}
