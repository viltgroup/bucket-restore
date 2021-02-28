module github.com/viltgroup/bucket-restore

go 1.16

replace github.com/aws/aws-sdk-go => github.com/roberth-k/aws-sdk-go v1.25.14-0.20200707084311-f2351a7ac473

require (
	cloud.google.com/go/storage v1.13.0
	github.com/aws/aws-sdk-go v1.37.14
	github.com/spf13/cobra v1.1.3
	google.golang.org/api v0.40.0
)
