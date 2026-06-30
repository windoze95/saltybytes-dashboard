package sgsync

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type mockEC2 struct {
	groups     []ec2types.SecurityGroup
	authorized []string
	revoked    []string
}

func (m *mockEC2) DescribeSecurityGroups(_ context.Context, _ *ec2.DescribeSecurityGroupsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{SecurityGroups: m.groups}, nil
}

func (m *mockEC2) AuthorizeSecurityGroupIngress(_ context.Context, in *ec2.AuthorizeSecurityGroupIngressInput, _ ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	for _, perm := range in.IpPermissions {
		for _, r := range perm.IpRanges {
			m.authorized = append(m.authorized, *r.CidrIp)
		}
	}
	return &ec2.AuthorizeSecurityGroupIngressOutput{}, nil
}

func (m *mockEC2) RevokeSecurityGroupIngress(_ context.Context, in *ec2.RevokeSecurityGroupIngressInput, _ ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	for _, perm := range in.IpPermissions {
		for _, r := range perm.IpRanges {
			m.revoked = append(m.revoked, *r.CidrIp)
		}
	}
	return &ec2.RevokeSecurityGroupIngressOutput{}, nil
}

func ipRange(cidr, desc string) ec2types.IpRange {
	r := ec2types.IpRange{CidrIp: aws.String(cidr)}
	if desc != "" {
		r.Description = aws.String(desc)
	}
	return r
}

func sgWith(ranges ...ec2types.IpRange) []ec2types.SecurityGroup {
	return []ec2types.SecurityGroup{{
		IpPermissions: []ec2types.IpPermission{{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(5432),
			ToPort:     aws.Int32(5432),
			IpRanges:   ranges,
		}},
	}}
}

const testHost = "longagoin.example.com"

func resolveTo(ip string) func(string) (string, error) {
	return func(string) (string, error) { return ip, nil }
}

func TestSyncOnce_AuthorizesMissingIP(t *testing.T) {
	cfg := Config{SecurityGroupID: "sg-1", DDNSHostname: testHost, Port: 5432}
	m := &mockEC2{groups: sgWith(ipRange("9.9.9.9/32", "unrelated rule"))}

	if err := syncOnce(context.Background(), m, cfg, resolveTo("1.2.3.4")); err != nil {
		t.Fatalf("syncOnce: %v", err)
	}
	if len(m.authorized) != 1 || m.authorized[0] != "1.2.3.4/32" {
		t.Errorf("authorized = %v, want [1.2.3.4/32]", m.authorized)
	}
	if len(m.revoked) != 0 {
		t.Errorf("revoked = %v, want none (unrelated rule must be left alone)", m.revoked)
	}
}

func TestSyncOnce_NoopWhenPresent(t *testing.T) {
	cfg := Config{SecurityGroupID: "sg-1", DDNSHostname: testHost, Port: 5432}
	m := &mockEC2{groups: sgWith(ipRange("1.2.3.4/32", ruleDescription(testHost)))}

	if err := syncOnce(context.Background(), m, cfg, resolveTo("1.2.3.4")); err != nil {
		t.Fatalf("syncOnce: %v", err)
	}
	if len(m.authorized) != 0 || len(m.revoked) != 0 {
		t.Errorf("expected no-op, got authorized=%v revoked=%v", m.authorized, m.revoked)
	}
}

func TestSyncOnce_RotatesStaleIP(t *testing.T) {
	cfg := Config{SecurityGroupID: "sg-1", DDNSHostname: testHost, Port: 5432}
	m := &mockEC2{groups: sgWith(
		ipRange("5.5.5.5/32", ruleDescription(testHost)), // our rule, old IP → revoke
		ipRange("8.8.8.8/32", "the api security group"),  // unrelated → keep
	)}

	if err := syncOnce(context.Background(), m, cfg, resolveTo("1.2.3.4")); err != nil {
		t.Fatalf("syncOnce: %v", err)
	}
	if len(m.authorized) != 1 || m.authorized[0] != "1.2.3.4/32" {
		t.Errorf("authorized = %v, want [1.2.3.4/32]", m.authorized)
	}
	if len(m.revoked) != 1 || m.revoked[0] != "5.5.5.5/32" {
		t.Errorf("revoked = %v, want [5.5.5.5/32] (and NOT the unrelated 8.8.8.8)", m.revoked)
	}
}
