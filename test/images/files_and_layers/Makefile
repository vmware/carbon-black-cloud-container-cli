build:
	# Build the helper binary
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build  \
			-o source/app \
			source/_app/main.go
	# Build the docker image
	docker build ./source -t files_and_layers_for_testing

export: build
	# Export the final image as a tar file for easier testing
	docker save files_and_layers_for_testing -o files_and_layers.tar