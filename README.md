Ethereum Kurtosis Lambda
=====================
This repo contains a [Kurtosis Lambda](https://docs.kurtosistech.com/lambdas.html) for starting a private Ethereum network. It is published to Dockerhub [here](https://hub.docker.com/repository/docker/kurtosistech/ethereum-kurtosis-lambda/). 

You can run this inside the [Kurtosis sandbox](https://docs.kurtosistech.com/sandbox.html) like so:

```javascript
loadLambdaResult = await networkCtx.loadLambda("eth-lambda", "kurtosistech/ethereum-kurtosis-lambda:0.2.1", "{}")
lambdaCtx = loadLambdaResult.value
executeResult = await lambdaCtx.execute("{}")
executeResultObj = JSON.parse(executeResult.value)
console.log(executeResultObj)
```

Once loaded, you can use libraries like `web3js` to interact with the Ethereum nodes.
