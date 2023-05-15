# promgo

A simple Cloudant Prometheus client that polls a Cloudant account for information
and publishes them in a Prometheus-consumable format on a `/metrics` endpoint.

## Configuring

Expects environment variables to be supplied according to [this documentation](https://cloud.ibm.com/apidocs/cloudant?code=go#authentication-with-external-configuration). e.g.

```sh
export CLOUDANT_URL="https://myservice.cloudant.com"
export CLOUDANT_APIKEY="my_IAM_API_KEY"
```

## Running locally

```sh
go run ./cmd/couchmonitor
```

## Running in Docker

First we turn this repo into a Docker image:
```sh
# build a docker image
docker build -t couchmonitor .
```

Then we can can spin up a container, exposing its port `8080` to our port `8080`:

```sh
# run it with credentials as environment variables
docker run \
  -e CLOUDANT_URL="$CLOUDANT_URL" -e CLOUDANT_APIKEY="$CLOUDANT_APIKEY" \
  -i \
  -p 8080:8080 \
  couchmonitor:latest
```
