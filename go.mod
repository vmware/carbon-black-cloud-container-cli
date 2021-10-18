module github.com/vmware/carbon-black-cloud-container-cli

go 1.16

require (
	github.com/anchore/stereoscope v0.0.0-20210413221244-d577f30b19e6
	github.com/anchore/syft v0.15.1
	github.com/containers/image/v5 v5.16.1
	github.com/docker/docker v20.10.8+incompatible
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/gookit/color v1.2.7
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.2.1
	github.com/manifoldco/promptui v0.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/wagoodman/go-partybus v0.0.0-20200526224238-eb215533f07d
	github.com/wagoodman/go-progress v0.0.0-20200731105512-1020f39e6240
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	gotest.tools/v3 v3.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.15
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.11
	github.com/go-restruct/restruct => github.com/go-restruct/restruct v1.2.0-alpha
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/google/go-containerregistry => github.com/google/go-containerregistry v0.4.1
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.2
	github.com/spf13/viper => github.com/spf13/viper v1.7.1
)
