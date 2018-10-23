# aws-ip-checks

With an authenticated session to AWS, run `./aws-ip-check` to understand your IP usage across AWS regions.

Currently assumes your subnets have a `Name` tag set.

1. Query all VPCs
2. Get all Subnets for VPC
3. Parse CIDR Block and Available IPs
4. Calculate Used IPs from Address Count of CIDR - Available

```
./aws-ip-check --help
Usage of ./aws-ip-check:
  -default-vpc
    	Include Default VPC(s).
  -region string
    	AWS Region. (default "us-west-2")
```