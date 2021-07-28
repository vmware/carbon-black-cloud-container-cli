module github.com/vmware/carbon-black-cloud-container-cli

go 1.15

require (
	github.com/anchore/stereoscope v0.0.0-20210413221244-d577f30b19e6
	github.com/anchore/syft v0.15.1
	github.com/containers/image/v5 v5.10.5
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/gookit/color v1.2.7
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-version v1.2.1
	github.com/manifoldco/promptui v0.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/wagoodman/go-partybus v0.0.0-20200526224238-eb215533f07d
	github.com/wagoodman/go-progress v0.0.0-20200731105512-1020f39e6240
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.15
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/go-restruct/restruct => github.com/go-restruct/restruct v1.2.0-alpha
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/google/go-containerregistry => github.com/google/go-containerregistry v0.4.1
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc93
	github.com/spf13/viper => github.com/spf13/viper v1.7.1
)
