module github.com/kurtosis-tech/ethereum-kurtosis-module

go 1.15

replace (
	github.com/kurtosis-tech/kurtosis-module-api-lib/golang => ../../kurtosis-module-api-lib/golang
)

require (
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20220915102629-0338bee14039
	github.com/kurtosis-tech/kurtosis-module-api-lib/golang v0.0.0-20220907164014-a41ac0f972e2
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/sirupsen/logrus v1.8.1
)
