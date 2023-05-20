package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"os"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InfrastructureStackProps struct {
	awscdk.StackProps
}

// TODO VPC, subnets, security group
// TODO EKS
// TODO load-balancer
// TODO Helm charts

func NewInfrastructureStack(scope constructs.Construct, id string, props *InfrastructureStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	_ = awsec2.NewVpc(stack, jsii.String("charity-platform"), &awsec2.VpcProps{
		VpcName:            jsii.String("charity-platform"),
		IpAddresses:        awsec2.IpAddresses_Cidr(jsii.String("172.31.0.0/16")),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             jsii.Number(2),
		NatGateways:        jsii.Number(2),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:                jsii.String("PublicK8s"),
				MapPublicIpOnLaunch: jsii.Bool(true),
				CidrMask:            jsii.Number(16),
				SubnetType:          awsec2.SubnetType_PUBLIC,
			},
			{
				Name:       jsii.String("PrivateK8s"),
				CidrMask:   jsii.Number(16),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		},
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfrastructureStack(app, "CharityPlatformStack", &InfrastructureStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
