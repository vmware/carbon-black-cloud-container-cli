# Auth

Extract the username and password from docker config json for the appropriate registry.
Returns empty strings for anonymous authentication when a registry entry isn't found.

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o auth main.go
```

## Run

```sh
./auth username|password registry|reference
```
