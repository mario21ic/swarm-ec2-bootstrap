package main

import (
	//"io"
	"fmt"
	"os"
	"log"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

func check(e error) {

	if e != nil {
		panic(e)
	}
}

func describe_nodes(role string, states []string, leader string) []string {
	sess := session.Must(session.NewSession())

	//awsRegion := "us-east-1"
	awsRegion := "us-west-2"
	svc := ec2.New(sess, &aws.Config{Region: aws.String(awsRegion)})
	//fmt.Printf("listing instances with tag %v in: %v\n", nameFilter, awsRegion)
	for _, stat := range states {
		fmt.Println("state: " + stat)
	}

	filters := []*ec2.Filter{
		{
			Name: aws.String("tag:Name"),
			Values: []*string{aws.String(role)},
		},
		{
			Name: aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
		{
			Name: aws.String("tag:Init"),
			Values: []*string{
				aws.String(strings.Join([]string{leader}, "")),
			},
		},
	}
	params := &ec2.DescribeInstancesInput{Filters: filters}
	resp, err := svc.DescribeInstances(params)

	if err != nil {
		fmt.Println("There was an error listing instances in", awsRegion, err.Error())
		log.Fatal(err.Error())
	}

	var instancesIds []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instancesIds = append(instancesIds,  *instance.InstanceId)
		}
	}
	//fmt.Println(instancesIds)

	return instancesIds
}

func main() {

	// Obteniendo variables de entorno
	role := os.Getenv("ROLE")
	//bm_env := os.Getenv("BMENV")
	//current_instance := os.Getenv("INSTANCE")

	fmt.Println(role)
	//fmt.Println(bm_env)
	//fmt.Println(current_instance)

	// Iniciando cli docker
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	check(err)
	cli.ContainerList(ctx, types.ContainerListOptions{})


	managers_running := describe_nodes("manager", []string{"running"}, "true")
	fmt.Printf("%+v\n", managers_running)

	managers_replaced := describe_nodes("manager", []string{"shutting-down", "stopped", "terminated", "stopping"}, "true")
	fmt.Printf("%+v\n", managers_replaced)

	worker_replaced := describe_nodes("worker", []string{"shutting-down", "stopped", "terminated", "stopping"}, "false")
	fmt.Printf("%+v\n", worker_replaced)
}
