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

const (
	LBCServiceAccountName = "aws-load-balancer-controller"
	arnPrefix             = "aws"
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

	lbcPolicy := awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		AssignSids: jsii.Bool(true),
		Statements: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("iam:CreateServiceLinkedRole"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"StringEquals": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S0"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"iam:AWSServiceName": "elasticloadbalancing.amazonaws.com",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:DescribeAccountAttributes"),
					jsii.String("ec2:DescribeAddresses"),
					jsii.String("ec2:DescribeAvailabilityZones"),
					jsii.String("ec2:DescribeInternetGateways"),
					jsii.String("ec2:DescribeVpcs"),
					jsii.String("ec2:DescribeVpcPeeringConnections"),
					jsii.String("ec2:DescribeSubnets"),
					jsii.String("ec2:DescribeSecurityGroups"),
					jsii.String("ec2:DescribeInstances"),
					jsii.String("ec2:DescribeNetworkInterfaces"),
					jsii.String("ec2:DescribeTags"),
					jsii.String("ec2:GetCoipPoolUsage"),
					jsii.String("ec2:DescribeCoipPools"),
					jsii.String("elasticloadbalancing:DescribeLoadBalancers"),
					jsii.String("elasticloadbalancing:DescribeLoadBalancerAttributes"),
					jsii.String("elasticloadbalancing:DescribeListeners"),
					jsii.String("elasticloadbalancing:DescribeListenerCertificates"),
					jsii.String("elasticloadbalancing:DescribeSSLPolicies"),
					jsii.String("elasticloadbalancing:DescribeRules"),
					jsii.String("elasticloadbalancing:DescribeTargetGroups"),
					jsii.String("elasticloadbalancing:DescribeTargetGroupAttributes"),
					jsii.String("elasticloadbalancing:DescribeTargetHealth"),
					jsii.String("elasticloadbalancing:DescribeTags"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("cognito-idp:DescribeUserPoolClient"),
					jsii.String("acm:ListCertificates"),
					jsii.String("acm:DescribeCertificate"),
					jsii.String("iam:ListServerCertificates"),
					jsii.String("iam:GetServerCertificate"),
					jsii.String("waf-regional:GetWebACL"),
					jsii.String("waf-regional:GetWebACLForResource"),
					jsii.String("waf-regional:AssociateWebACL"),
					jsii.String("waf-regional:DisassociateWebACL"),
					jsii.String("wafv2:GetWebACL"),
					jsii.String("wafv2:GetWebACLForResource"),
					jsii.String("wafv2:AssociateWebACL"),
					jsii.String("wafv2:DisassociateWebACL"),
					jsii.String("shield:GetSubscriptionState"),
					jsii.String("shield:DescribeProtection"),
					jsii.String("shield:CreateProtection"),
					jsii.String("shield:DeleteProtection"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:AuthorizeSecurityGroupIngress"),
					jsii.String("ec2:RevokeSecurityGroupIngress"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateSecurityGroup"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:" + arnPrefix + ":ec2:*:*:security-group/*"),
				},
				Conditions: &map[string]interface{}{
					"StringEquals": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S1"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"ec2:CreateAction": "CreateSecurityGroup",
						},
					}),
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S2"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:CreateTags"),
					jsii.String("ec2:DeleteTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:" + arnPrefix + ":ec2:*:*:security-group/*"),
				},
				Conditions: &map[string]interface{}{
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S3"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/elbv2.k8s.aws/cluster":  "true",
							"aws:ResourceTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("ec2:AuthorizeSecurityGroupIngress"),
					jsii.String("ec2:RevokeSecurityGroupIngress"),
					jsii.String("ec2:DeleteSecurityGroup"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S4"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:ResourceTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:CreateLoadBalancer"),
					jsii.String("elasticloadbalancing:CreateTargetGroup"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S5"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:CreateListener"),
					jsii.String("elasticloadbalancing:DeleteListener"),
					jsii.String("elasticloadbalancing:CreateRule"),
					jsii.String("elasticloadbalancing:DeleteRule"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:AddTags"),
					jsii.String("elasticloadbalancing:RemoveTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:targetgroup/*/*"),
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:loadbalancer/net/*/*"),
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:loadbalancer/app/*/*"),
				},
				Conditions: &map[string]interface{}{
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S6"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:RequestTag/elbv2.k8s.aws/cluster":  "true",
							"aws:ResourceTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:AddTags"),
					jsii.String("elasticloadbalancing:RemoveTags"),
				},
				Resources: &[]*string{
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:listener/net/*/*/*"),
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:listener/app/*/*/*"),
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:listener-rule/net/*/*/*"),
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:listener-rule/app/*/*/*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:ModifyLoadBalancerAttributes"),
					jsii.String("elasticloadbalancing:SetIpAddressType"),
					jsii.String("elasticloadbalancing:SetSecurityGroups"),
					jsii.String("elasticloadbalancing:SetSubnets"),
					jsii.String("elasticloadbalancing:DeleteLoadBalancer"),
					jsii.String("elasticloadbalancing:ModifyTargetGroup"),
					jsii.String("elasticloadbalancing:ModifyTargetGroupAttributes"),
					jsii.String("elasticloadbalancing:DeleteTargetGroup"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
				Conditions: &map[string]interface{}{
					"Null": awscdk.NewCfnJson(stack, jsii.String("CfnJson-LBC-S7"), &awscdk.CfnJsonProps{
						Value: map[string]string{
							"aws:ResourceTag/elbv2.k8s.aws/cluster": "false",
						},
					}),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:RegisterTargets"),
					jsii.String("elasticloadbalancing:DeregisterTargets"),
				},
				Resources: &[]*string{
					jsii.String("arn:" + arnPrefix + ":elasticloadbalancing:*:*:targetgroup/*/*"),
				},
			}),
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticloadbalancing:SetWebAcl"),
					jsii.String("elasticloadbalancing:ModifyListener"),
					jsii.String("elasticloadbalancing:AddListenerCertificates"),
					jsii.String("elasticloadbalancing:RemoveListenerCertificates"),
					jsii.String("elasticloadbalancing:ModifyRule"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
		},
	})

	lbcSa := awseks.NewServiceAccount(stack, jsii.String("AWSLoadBalancerControllerSA"), &awseks.ServiceAccountProps{
		Name:      jsii.String(LBCServiceAccountName),
		Cluster:   cluster,
		Namespace: jsii.String("kube-system"),
	})

	awsiam.NewPolicy(stack, jsii.String("AWSLoadBalancerControllerPolicy"), &awsiam.PolicyProps{
		Document:   lbcPolicy,
		PolicyName: jsii.String(*stack.StackName() + "-AWSLoadBalancerControllerIAMPolicy"),
		Roles: &[]awsiam.IRole{
			lbcSa.Role(),
		},
	})

	// https://github.com/kubernetes-sigs/aws-load-balancer-controller/tree/main/helm/aws-load-balancer-controller
	// TODO: --set image.repository=
	// TODO: https://docs.aws.amazon.com/eks/latest/userguide/add-ons-images.html
	lbcChart := awseks.NewHelmChart(stack, jsii.String("AWSLoadBalancerControllerChart"), &awseks.HelmChartProps{
		Cluster:         cluster,
		Repository:      jsii.String("https://aws.github.io/eks-charts"),
		Release:         jsii.String("aws-load-balancer-controller"),
		Chart:           jsii.String("aws-load-balancer-controller"),
		Namespace:       jsii.String("kube-system"),
		CreateNamespace: jsii.Bool(true),
		Wait:            jsii.Bool(true),
		Version:         jsii.String("1.4.5"),
		Values: &map[string]interface{}{
			"clusterName": *cluster.ClusterName(),
			"defaultTags": map[string]string{
				"eks:cluster-name": *cluster.ClusterName(),
			},
			"region": stack.Region(),
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "200m",
					"memory": "160Mi",
				},
				"requests": map[string]any{
					"cpu":    "100m",
					"memory": "80Mi",
				},
			},
			"podDisruptionBudget": map[string]any{
				"maxUnavailable": 1,
			},
			"priorityClassName": "system-cluster-critical",
			"hostNetwork":       jsii.Bool(true),
			"vpcId":             cluster.Vpc().VpcId(),
			"serviceAccount": map[string]interface{}{
				"create": false,
				"name":   LBCServiceAccountName,
				/*
					"annotations": map[string]interface{}{
						"eks.amazonaws.com/sts-regional-endpoints": jsii.Bool(true),
					},
				*/
			},
			"nodeSelector": map[string]string{
				"kubernetes.io/os": "linux",
			},
		},
	})
	lbcChart.Node().AddDependency(lbcSa)
	lbcChart.Node().AddDependency(lbcSa)

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
