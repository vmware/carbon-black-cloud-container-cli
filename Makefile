version := ${CLI_VERSION}
build_date := `date +%Y/%m/%d`
build_tags := "containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs"

.PHONY: dep build test

dep:
	go mod download

test:
	go test ./... -tags=${build_tags} -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

build:
	echo "Building CLI with version: $(version), build date: $(build_date)"

	# build binary for mac osx
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
		go build  \
			-tags=${build_tags} \
			-o bin/darwin/cbctl \
			-ldflags "-X 'gitlab.bit9.local/octarine/cbctl/internal/version.version=${version}' \
				-X 'gitlab.bit9.local/octarine/cbctl/internal/version.buildDate=${build_date}'" \
			main.go

	# build binary for linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build  \
			-tags=${build_tags} \
			-o bin/linux/cbctl \
			-ldflags "-X 'gitlab.bit9.local/octarine/cbctl/internal/version.version=${version}' \
				-X 'gitlab.bit9.local/octarine/cbctl/internal/version.buildDate=${build_date}'" \
			main.go



