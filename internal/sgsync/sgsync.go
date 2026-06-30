// Package sgsync keeps this host's (dynamic) public IP whitelisted on an AWS
// security group, so a self-hosted dashboard on a changing ISP IP can always
// reach RDS. It resolves a DDNS hostname to the current IP and authorizes it on
// a port, revoking its own stale rules. Disabled unless configured.
package sgsync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Config controls the security-group auto-sync.
type Config struct {
	Region          string
	SecurityGroupID string
	DDNSHostname    string
	Port            int
	Interval        time.Duration
}

// Enabled reports whether the sync is configured (both the SG and hostname set).
func (c Config) Enabled() bool {
	return c.SecurityGroupID != "" && c.DDNSHostname != ""
}

// FromEnv builds config from env. No-op unless SGSYNC_SECURITY_GROUP_ID and
// SGSYNC_DDNS_HOSTNAME are both set. AWS creds come from the standard chain
// (AWS_ACCESS_KEY_ID/SECRET or an instance role).
func FromEnv() Config {
	port := 5432
	if v := os.Getenv("SGSYNC_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			port = n
		}
	}
	interval := 5 * time.Minute
	if v := os.Getenv("SGSYNC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("SGSYNC_REGION")
	}
	return Config{
		Region:          region,
		SecurityGroupID: os.Getenv("SGSYNC_SECURITY_GROUP_ID"),
		DDNSHostname:    os.Getenv("SGSYNC_DDNS_HOSTNAME"),
		Port:            port,
		Interval:        interval,
	}
}

// ec2API is the subset of the EC2 client this package needs (mockable in tests).
type ec2API interface {
	DescribeSecurityGroups(ctx context.Context, in *ec2.DescribeSecurityGroupsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	AuthorizeSecurityGroupIngress(ctx context.Context, in *ec2.AuthorizeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupIngress(ctx context.Context, in *ec2.RevokeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
}

func resolveIPv4(host string) (string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			return v4.String(), nil
		}
	}
	return "", fmt.Errorf("no IPv4 address for %q", host)
}

// ruleDescription tags rules this syncer manages so it only ever revokes its
// own / hostname-tagged entries — never the API's SG rule or unrelated rules.
func ruleDescription(host string) string {
	return "saltybytes-dashboard sgsync " + host
}

// syncOnce ensures the security group allows the host's current IP on the port,
// and revokes any stale rules previously added for the same hostname. Testable
// core: the EC2 client and DNS resolver are injected.
func syncOnce(ctx context.Context, client ec2API, cfg Config, resolve func(string) (string, error)) error {
	ip, err := resolve(cfg.DDNSHostname)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", cfg.DDNSHostname, err)
	}
	desired := ip + "/32"

	out, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{cfg.SecurityGroupID},
	})
	if err != nil {
		return fmt.Errorf("describe security group: %w", err)
	}

	allowed := false
	var stale []string // CIDRs of prior rules for this hostname pointing elsewhere
	for _, sg := range out.SecurityGroups {
		for _, perm := range sg.IpPermissions {
			if perm.FromPort == nil || *perm.FromPort != int32(cfg.Port) {
				continue
			}
			for _, r := range perm.IpRanges {
				if r.CidrIp == nil {
					continue
				}
				if *r.CidrIp == desired {
					allowed = true
					continue
				}
				if r.Description != nil && strings.Contains(*r.Description, cfg.DDNSHostname) {
					stale = append(stale, *r.CidrIp)
				}
			}
		}
	}

	if !allowed {
		if _, err := client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(cfg.SecurityGroupID),
			IpPermissions: []ec2types.IpPermission{{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(int32(cfg.Port)),
				ToPort:     aws.Int32(int32(cfg.Port)),
				IpRanges: []ec2types.IpRange{{
					CidrIp:      aws.String(desired),
					Description: aws.String(ruleDescription(cfg.DDNSHostname)),
				}},
			}},
		}); err != nil {
			return fmt.Errorf("authorize %s: %w", desired, err)
		}
		log.Printf("sgsync: authorized %s on %s:%d", desired, cfg.SecurityGroupID, cfg.Port)
	}

	for _, cidr := range stale {
		if _, err := client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId: aws.String(cfg.SecurityGroupID),
			IpPermissions: []ec2types.IpPermission{{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(int32(cfg.Port)),
				ToPort:     aws.Int32(int32(cfg.Port)),
				IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(cidr)}},
			}},
		}); err != nil {
			log.Printf("sgsync: failed to revoke stale %s: %v", cidr, err)
			continue
		}
		log.Printf("sgsync: revoked stale %s", cidr)
	}
	return nil
}

// SyncOnce performs a single sync against the real EC2 API.
func SyncOnce(ctx context.Context, cfg Config) error {
	if !cfg.Enabled() {
		return errors.New("sgsync: not configured")
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	return syncOnce(ctx, ec2.NewFromConfig(awsCfg), cfg, resolveIPv4)
}

// Run syncs once synchronously (so the current IP is whitelisted before the
// caller connects to the DB), then keeps syncing on an interval in the
// background. Best-effort — failures are logged, never fatal. No-op when
// unconfigured.
func Run(ctx context.Context, cfg Config) {
	if !cfg.Enabled() {
		log.Printf("sgsync: disabled (set SGSYNC_SECURITY_GROUP_ID + SGSYNC_DDNS_HOSTNAME to enable)")
		return
	}
	if err := SyncOnce(ctx, cfg); err != nil {
		log.Printf("sgsync: initial sync failed (continuing): %v", err)
	}
	go func() {
		t := time.NewTicker(cfg.Interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := SyncOnce(ctx, cfg); err != nil {
					log.Printf("sgsync: periodic sync failed: %v", err)
				}
			}
		}
	}()
}
