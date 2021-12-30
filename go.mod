module dev.nimak.link/s3-copy-controller

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.11.2
	github.com/aws/aws-sdk-go-v2/config v1.11.0
	github.com/aws/aws-sdk-go-v2/credentials v1.6.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.21.0
	github.com/go-ini/ini v1.66.2
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
