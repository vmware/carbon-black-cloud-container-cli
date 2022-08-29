version:=${CLI_VERSION}
build_date:=`date +%Y/%m/%d`

build:
	echo "Building CLI with version: $(version), build date: $(build_date)"

	# build binary for mac osx
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
		go build  \
			-tags="containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs" \
			-o bin/darwin/cbctl \
			-ldflags "-X 'gitlab.bit9.local/octarine/cbctl/internal/version.version=${version}' \
				-X 'gitlab.bit9.local/octarine/cbctl/internal/version.buildDate=${build_date}'" \
			main.go

	# build binary for linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build  \
			-tags="containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs" \
			-o bin/linux/cbctl \
			-ldflags "-X 'gitlab.bit9.local/octarine/cbctl/internal/version.version=${version}' \
				-X 'gitlab.bit9.local/octarine/cbctl/internal/version.buildDate=${build_date}'" \
			main.go



