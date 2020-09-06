module github.com/omakeno/kubectl-kbkb

go 1.14

require (
	github.com/omakeno/bashoverwriter v0.1.1
	github.com/omakeno/kbkb v0.1.8
	github.com/spf13/cobra v1.0.0
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.0.0-20200626130735-db5293afc7bf
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20200626130448-f849118f70f6
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20200626130251-3b98a76529ae
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20200626132723-2eb52e397b36
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200626130735-db5293afc7bf
)
