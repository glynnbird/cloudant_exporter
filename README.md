# promgo

A simple Cloudant Prometheus client that polls a Cloudant account for information
and publishes them in a Prometheus-consumable format on a `/metrics` endpoint.

## Configuring

Expects environment variables to be supplied according to [this documentation](https://cloud.ibm.com/apidocs/cloudant?code=go#authentication-with-external-configuration). e.g.

```sh
export CLOUDANT_URL="https://myservice.cloudant.com"
export CLOUDANT_APIKEY="my_IAM_API_KEY"
```

## Running

```sh
go run cmd/couchmonitor/main.go 
```