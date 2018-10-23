package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var region string
	var defaultVPC bool
	flag.BoolVar(&defaultVPC, "default-vpc", false, "Include Default VPC(s).")
	flag.StringVar(&region, "region", "us-west-2", "AWS Region.")
	flag.Parse()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Subnet", "Used", "Available", "Size", "Sum Used", "Sum Max"})
	var sumUsed uint64
	var sumMax uint64

	sess, err := session.NewSession(&aws.Config{
		Region: &region},
	)

	svc := ec2.New(sess)
	vpcInput := ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("isDefault"),
				Values: []*string{
					aws.String(strconv.FormatBool(defaultVPC)),
				},
			},
		}}
	result, err := svc.DescribeVpcs(&vpcInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	for _, v := range result.Vpcs {
		input := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						v.VpcId,
					},
				},
			}}
		result, err := svc.DescribeSubnets(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}
		for _, v := range result.Subnets {
			used, total := parseResults(*v.CidrBlock, *v.AvailableIpAddressCount)
			sumUsed += used
			sumMax += total
			var tag string
			for _, v := range v.Tags {
				if *v.Key == "Name" {
					tag = *v.Value
				}
			}
			table.Append([]string{tag, *v.CidrBlock, strconv.FormatUint(used, 10), strconv.FormatInt(*v.AvailableIpAddressCount, 10), strconv.FormatUint(total, 10), strconv.FormatUint(sumUsed, 10), strconv.FormatUint(sumMax, 10)})
		}
	}
	table.Render()
}

func parseResults(CidrBlock string, AvailableIpAddressCount int64) (uint64, uint64) {
	_, netCIDR, err := net.ParseCIDR(CidrBlock)
	if err != nil {
		fmt.Errorf("Failed to parse CIDR: %v", err)
	}
	count := AddressCount(netCIDR)

	return (count - uint64(AvailableIpAddressCount)), count
}

// AddressCount returns the number of distinct host addresses within the given
// CIDR range.
//
// Since the result is a uint64, this function returns meaningful information
// only for IPv4 ranges and IPv6 ranges with a prefix size of at least 65.
func AddressCount(network *net.IPNet) uint64 {
	prefixLen, bits := network.Mask.Size()
	return 1 << (uint64(bits) - uint64(prefixLen))
}
