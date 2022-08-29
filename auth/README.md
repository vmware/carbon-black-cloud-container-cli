# Auth

Extract the username and password from docker config json for the appropriate registry.
Returns empty strings for anonymous authentication when a registry entry isn't found.

Vendored from <https://gitlab.eng.vmware.com/vulnerability-scanning-enablement/snyk-scanner-integration/-/tree/main/images/scanner/auth>.

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o auth main.go
```

## Run

```sh
./auth username|password registry|reference
```
