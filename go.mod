module gitlab.bit9.local/octarine/cbctl

go 1.15

require (
	github.com/anchore/stereoscope v0.0.0-20220201190559-f162f1e96f45
	github.com/anchore/syft v0.37.5
	github.com/containers/image/v5 v5.10.5
	github.com/docker/docker v20.10.12+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/google/uuid v1.2.0 // indirect
	github.com/gookit/color v1.2.7
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-version v1.2.1
	github.com/manifoldco/promptui v0.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/wagoodman/go-partybus v0.0.0-20210627031916-db1f5573bbc5
	github.com/wagoodman/go-progress v0.0.0-20200731105512-1020f39e6240
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	gotest.tools/v3 v3.1.0 // indirect
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
