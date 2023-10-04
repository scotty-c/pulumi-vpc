package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi-awsx/sdk/go/awsx/ecr"
	ecrx "github.com/pulumi/pulumi-awsx/sdk/go/awsx/ecr"
	ecsx "github.com/pulumi/pulumi-awsx/sdk/go/awsx/ecs"
	lbx "github.com/pulumi/pulumi-awsx/sdk/go/awsx/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// cfg sets defaults for all the variables if there are no values in Pulumi.dev.yaml
		cfg := config.New(ctx, "")
		stackName := "go-demo"
		if param := cfg.Get("stackName"); param != "" {
			stackName = param
		}
		clusterName := "default"
		if param := cfg.Get("clusterName"); param != "" {
			clusterName = param
		}
		lbName := "go-demo"
		if param := cfg.Get("lbName"); param != "" {
			lbName = param
		}
		repoName := "go-demo"
		if param := cfg.Get("repoName"); param != "" {
			repoName = param
		}
		imageName := "go-demo"
		if param := cfg.Get("imageName"); param != "" {
			imageName = param
		}
		containerPort := 80
		if param := cfg.GetInt("containerPort"); param != 0 {
			containerPort = param
		}
		cpu := 512
		if param := cfg.GetInt("cpu"); param != 0 {
			cpu = param
		}
		memory := 128
		if param := cfg.GetInt("memory"); param != 0 {
			memory = param
		}

		// An ECS cluster to deploy into
		cluster, err := ecs.NewCluster(ctx, clusterName, nil)
		if err != nil {
			return err
		}

		// An ALB to serve the container endpoint to the internet
		loadbalancer, err := lbx.NewApplicationLoadBalancer(ctx, lbName, nil)
		if err != nil {
			return err
		}

		// An ECR repository to store our application's container image
		repo, err := ecrx.NewRepository(ctx, repoName, &ecrx.RepositoryArgs{
			ForceDelete: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Build and publish our application's container image from ./app to the ECR repository
		image, err := ecrx.NewImage(ctx, imageName, &ecr.ImageArgs{
			RepositoryUrl: repo.Url,
			Path:          pulumi.String("./app"),
		})
		if err != nil {
			return err
		}

		// Deploy an ECS Service on Fargate to host the application container
		_, err = ecsx.NewFargateService(ctx, stackName, &ecsx.FargateServiceArgs{
			Cluster:        cluster.Arn,
			AssignPublicIp: pulumi.Bool(true),
			TaskDefinitionArgs: &ecsx.FargateServiceTaskDefinitionArgs{
				Container: &ecsx.TaskDefinitionContainerDefinitionArgs{
					Image:     image.ImageUri,
					Cpu:       pulumi.Int(cpu),
					Memory:    pulumi.Int(memory),
					Essential: pulumi.Bool(true),
					PortMappings: ecsx.TaskDefinitionPortMappingArray{
						&ecsx.TaskDefinitionPortMappingArgs{
							ContainerPort: pulumi.Int(containerPort),
							TargetGroup:   loadbalancer.DefaultTargetGroup,
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// The URL at which the container's HTTP endpoint will be available
		ctx.Export("url", pulumi.Sprintf("http://%s", loadbalancer.LoadBalancer.DnsName()))
		return nil
	})
}
