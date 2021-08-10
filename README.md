# carbon-black-cloud-container-cli

_carbon-black-cloud-container-cli_ (also known as _cbctl_) is a CLI tool that can be used to scan any container-based 
images in the command line or in CI/CD pipelines.

## Get started

### Build the CLI binary

To Linux: 
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build  \
   -tags="containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs" \
   -ldflags "-X 'github.com/vmware/carbon-black-cloud-container-cli/internal/version.version=${version}' \
             -X 'github.com/vmware/carbon-black-cloud-container-cli/internal/version.buildDate=${build_date}'" \
   main.go
```

To MacOS:
```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
  go build  \
   -tags="containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs" \
   -ldflags "-X 'github.com/vmware/carbon-black-cloud-container-cli/internal/version.version=${version}' \
             -X 'github.com/vmware/carbon-black-cloud-container-cli/internal/version.buildDate=${build_date}'" \
   main.go
```

### CLI binary reference

The detailed usage of _cbctl_ can be found here: [Carbon Black Container CLI](https://developer.carbonblack.com/reference/carbon-black-cloud/container/latest/image-scanning-cli/)

## Package usage

We exposed bom generation and scan image functions for convenient image scanning process, you can follow the following 
steps to get started:

### Import the package

`$ go get -u github.com/vmware/carbon-black-cloud-container-cli`

### How to take use of CLI packages?

1. Create a pair of API ID & Key with `workloads.container.image` (CREATE and READ permissions) access level in Carbon Black Cloud console
2. Create a RegistryHandler for generating Software Bill of Materials (SBOM) from user's input: 
   1. Create RegistryHandler: `registryHandler := scan.NewRegistryHandler()`
   2. Get the SBOM (options can be checked below): `sbom, err := registryHandler.Generate(input, scan.Option)`
3. Create a ScanHandler for scanning vulnerabilities from SBOM:
   1. Create ScanHandler: `scanHandler := scan.NewScanHandler(<CBC_saasURL>, <CBC_orgKey>, <apiID>, <apiKey>, <sbom>)`;
   2. Scan the SBOM (options can be checked below): `scannedImage, err := scanHandler.Scan(scan.Option)`

#### Scan options
| Option Name | Type | Description |
| --- | --- | --- |
| ForceScan | bool | Force scan an image no matter it is scanned or not |
| Credential | string | The auth string used for login to registry, format: `USERNAME[:PASSWORD]` |
| FullTag | string | The tag set to override in the image |
| UseDockerDaemon | bool | Use docker daemon to pull the image |
| ShouldCleanup | bool | Delete the docker image pulled by docker (should only be used when `UserDockerDaemon` is `true`) |
| Timeout | int | The duration (second) for the scan |

## Contributing

Please follow [CONTRIBUTING.md](CONTRIBUTING.md)

## License

[Apache-2.0](LICENSE)