package main

import (
	"flag"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	layer "github.com/cdklabs/awscdk-kubectl-go/kubectlv26/v2"
	"os"
)

type InfrastructureStackProps struct {
	awscdk.StackProps
	CirdVPC                                                                                       *string
	AvailabilityZones                                                                             *float64
	NodeAutoScalingGroupMaxSize, NodeAutoScalingGroupDesiredCapacity, NodeAutoScalingGroupMinSize *float64
	KubernetesVersion                                                                             *string
}

type ClusterValues struct {
	Version                                                                                       *string
	VPC                                                                                           awsec2.Vpc
	NodeAutoScalingGroupMaxSize, NodeAutoScalingGroupDesiredCapacity, NodeAutoScalingGroupMinSize *float64
}

func NewInfrastructureStack(scope constructs.Construct, id string, props *InfrastructureStackProps) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	vpc, err := addVPC(stack, props.CirdVPC, props.AvailabilityZones)
	if err != nil {
		return nil, err
	}

	_ = addEKS(stack, ClusterValues{
		Version:                             props.KubernetesVersion,
		VPC:                                 vpc,
		NodeAutoScalingGroupMaxSize:         props.NodeAutoScalingGroupMaxSize,
		NodeAutoScalingGroupDesiredCapacity: props.NodeAutoScalingGroupDesiredCapacity,
		NodeAutoScalingGroupMinSize:         props.NodeAutoScalingGroupMinSize,
	})

	return stack, nil
}

func addVPC(stack awscdk.Stack, vpcCIDR *string, azs *float64) (awsec2.Vpc, error) {
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

func addEKS(stack awscdk.Stack, params ClusterValues) awseks.Cluster {
	clusterAdmin := awsiam.NewRole(stack, jsii.String("KubernetesAdminRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewAccountRootPrincipal(),
	})

	subnetSelector := &awsec2.SubnetSelection{}
	subnetSelector.Subnets = params.VPC.PrivateSubnets()

	kmsKey := awskms.NewKey(stack, jsii.String("KMSKey"), &awskms.KeyProps{
		Alias:             jsii.String("charity-platform"),
		Enabled:           jsii.Bool(true),
		KeySpec:           awskms.KeySpec_SYMMETRIC_DEFAULT,
		EnableKeyRotation: jsii.Bool(true),
		KeyUsage:          awskms.KeyUsage_ENCRYPT_DECRYPT,
	})

	cluster := awseks.NewCluster(stack, jsii.String("cluster"), &awseks.ClusterProps{
		Version:           awseks.KubernetesVersion_Of(params.Version),
		ClusterName:       jsii.String("charity-platform"),
		OutputClusterName: jsii.Bool(true),
		Vpc:               params.VPC,
		VpcSubnets: &[]*awsec2.SubnetSelection{
			subnetSelector,
		},
		KubectlLayer:         layer.NewKubectlV26Layer(stack, jsii.String("KubectlV26Layer")),
		SecretsEncryptionKey: kmsKey,
		ClusterLogging: &[]awseks.ClusterLoggingTypes{
			awseks.ClusterLoggingTypes_API,
			awseks.ClusterLoggingTypes_AUDIT,
			awseks.ClusterLoggingTypes_AUTHENTICATOR,
			awseks.ClusterLoggingTypes_CONTROLLER_MANAGER,
			awseks.ClusterLoggingTypes_SCHEDULER,
		},
		EndpointAccess:       awseks.EndpointAccess_PUBLIC_AND_PRIVATE(),
		MastersRole:          clusterAdmin,
		OutputMastersRoleArn: jsii.Bool(true),
	})
	cluster.AddAutoScalingGroupCapacity(jsii.String("charity-platform-au-group"), &awseks.AutoScalingGroupCapacityOptions{
		AllowAllOutbound:     jsii.Bool(true),
		AutoScalingGroupName: jsii.String("charity-platform-au-group"),
		DesiredCapacity:      params.NodeAutoScalingGroupDesiredCapacity,
		MaxCapacity:          params.NodeAutoScalingGroupMaxSize,
		MinCapacity:          params.NodeAutoScalingGroupMinSize,
		InstanceType:         awsec2.NewInstanceType(jsii.String("t2.micro")),
		BootstrapEnabled:     jsii.Bool(true),
		MapRole:              jsii.Bool(true),
	})

	_ = cluster.AddHelmChart(jsii.String("metrics-server"), &awseks.HelmChartOptions{
		Chart:           jsii.String("metrics-server"),
		CreateNamespace: jsii.Bool(true),
		Namespace:       jsii.String("metrics-server"),
		Release:         jsii.String("metrics-server"),
		Repository:      jsii.String("https://kubernetes-sigs.github.io/metrics-server/"),
		Values: &map[string]any{
			"hostNetwork": map[string]any{
				"enabled": true,
			},
			"metrics": map[string]any{
				"enabled": true,
			},
			"containerPort": 4443,
		},
		Version: jsii.String("3.10.0"),
	})

	return cluster
}

var (
	cidr                                = flag.String("cidr", "172.31.0.0/16", "CIDR value for a VPC")
	azs                                 = flag.Float64("azs", 2, "Availability zone count for a VPC")
	nodeAutoScalingGroupMaxSize         = flag.Float64("ng-max-size", 10, "Maximum amount of Nodes in cluster")
	nodeAutoScalingGroupDesiredCapacity = flag.Float64("ng-desired-size", 2, "Desired amount of Nodes in cluster")
	nodeAutoScalingGroupMinSize         = flag.Float64("ng-minimum-size", 1, "Minimum amount of Nodes in cluster")
	kubernetesVersion                   = flag.String("k8s-version", "1.26", "Version of Kubernetes")
)

func main() {
	defer jsii.Close()
	flag.Parse()

	app := awscdk.NewApp(nil)

	_, err := NewInfrastructureStack(app, "CharityPlatformStack", &InfrastructureStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		CirdVPC:                             cidr,
		AvailabilityZones:                   azs,
		NodeAutoScalingGroupMaxSize:         nodeAutoScalingGroupMaxSize,
		NodeAutoScalingGroupDesiredCapacity: nodeAutoScalingGroupDesiredCapacity,
		NodeAutoScalingGroupMinSize:         nodeAutoScalingGroupMinSize,
		KubernetesVersion:                   kubernetesVersion,
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
