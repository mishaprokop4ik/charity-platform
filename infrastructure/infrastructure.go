package main

import (
	"flag"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"os"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InfrastructureStackProps struct {
	awscdk.StackProps
	CirdVPC                                                                                       *string
	AvailabilityZones                                                                             *float64
	NodeAutoScalingGroupMaxSize, NodeAutoScalingGroupDesiredCapacity, NodeAutoScalingGroupMinSize *uint
}

// TODO EKS
// TODO load-balancer
// TODO Helm charts

func NewInfrastructureStack(scope constructs.Construct, id string, props *InfrastructureStackProps) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	if _, err := addVPC(stack, props.CirdVPC, props.AvailabilityZones); err != nil {
		return nil, err
	}

	return stack, nil
}

func addVPC(stack awscdk.Stack, vpcCIDR *string, azs *float64) (awsec2.Vpc, error) {
	//cirdMask, err := strconv.Atoi(strings.Split(*vpcCIDR, "/")[1])
	//if err != nil {
	//	log.Printf("got error %s", err.Error())
	//	return nil, err
	//}
	vpc := awsec2.NewVpc(stack, jsii.String("charity-platform"), &awsec2.VpcProps{
		VpcName:            jsii.String("charity-platform"),
		IpAddresses:        awsec2.IpAddresses_Cidr(vpcCIDR),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             azs,
		NatGateways:        azs,
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:                jsii.String("PublicK8s"),
				MapPublicIpOnLaunch: jsii.Bool(true),
				CidrMask:            jsii.Number(float64(24)),
				SubnetType:          awsec2.SubnetType_PUBLIC,
			},
			{
				Name:       jsii.String("PrivateK8s"),
				CidrMask:   jsii.Number(float64(24)),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		},
	})

	return vpc, nil
}

func addEKS(stack awscdk.Stack) {

}

var (
	cidr = flag.String("cidr", "172.31.0.0/16", "CIDR value for a VPC")
	azs  = flag.Float64("azs", 2, "Availability zone count for a VPC")
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	_, err := NewInfrastructureStack(app, "CharityPlatformStack", &InfrastructureStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		CirdVPC:           cidr,
		AvailabilityZones: azs,
	})
	if err != nil {
		return
	}

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
