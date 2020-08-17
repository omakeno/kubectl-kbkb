module github.com/omakeno/kubectl-kbkb

go 1.14

require (
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/omakeno/kbkb v0.1.1
	github.com/spf13/cobra v1.0.0
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/api v0.18.8 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.0.0-20200626130735-db5293afc7bf
	k8s.io/klog/v2 v2.3.0 // indirect
	k8s.io/utils v0.0.0-20200815180417-3bc9d57fc792 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20200626130448-f849118f70f6
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20200626130251-3b98a76529ae
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20200626132723-2eb52e397b36
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200626130735-db5293afc7bf
)
