package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	r53sdk "github.com/aws/aws-sdk-go/service/route53"
	"github.com/rs/zerolog/log"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/apimachinery/pkg/util/sets"
)

const resolveAttempts = 100

func refresh(ctx context.Context, yc *ycsdk.SDK, r53 *r53sdk.Route53, gid string) {
	l := log.With().Str("groupId", gid).Logger()
	l.Info().Msg("refreshing state")

	req := instancegroup.ListInstanceGroupInstancesRequest{
		InstanceGroupId: gid,
		// NOTE: we do not expect large instance groups
		PageSize: 1000,
	}

	res, err := yc.InstanceGroup().InstanceGroup().ListInstances(ctx, &req)
	if err != nil {
		l.Err(err).Msg("unable to list instances")
		return
	}

	ips := make([]string, len(res.Instances))
	for i, h := range res.Instances {
		ifcfg := h.GetNetworkInterfaces()
		if len(ifcfg) < 1 {
			continue
		}
		ips[i] = ifcfg[0].PrimaryV4Address.OneToOneNat.Address
	}

	l.Info().Strs("ips", ips).Msg("received groupd members")

	rec, err := net.LookupHost(*args.Name)
	if err != nil {
		l.Err(err).Msg("LookupHost failed")
		dnsErr, ok := err.(*net.DNSError)
		if !ok {
			return
		}
		if !dnsErr.IsNotFound {
			return
		}
	}

	l.Info().
		Str("name", *args.Name).
		Strs("reolve", rec).
		Msg("resolved name")

	if sets.NewString(ips...).Equal(sets.NewString(rec...)) {
		l.Info().Msg("instance group not changed")
		return
	}

	rs := make([]*route53.ResourceRecord, len(ips))
	for i, ip := range ips {
		rs[i] = &route53.ResourceRecord{
			Value: aws.String(ip),
		}
	}

	ttl := int64(*args.TTL)

	in := r53sdk.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(*args.Zone),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            args.Name,
						ResourceRecords: rs,
						TTL:             &ttl,
						Type:            aws.String("A"),
					},
				},
			},
			Comment: args.Comment,
		},
	}

	l.Info().Msg("updating DNS record")

	out, err := r53.ChangeResourceRecordSets(&in)
	if err != nil {
		l.Err(err).Msg("unable to change record")
		return
	}

	l = l.With().Str("changeId", *out.ChangeInfo.Id).Logger()

	wch := route53.GetChangeInput{Id: out.ChangeInfo.Id}

	l.Info().Msg("waiting for change")
	if err := r53.WaitUntilResourceRecordSetsChanged(&wch); err != nil {
		l.Err(err).Msg("change failed")
		return
	}

	l.Info().Msg("change applied")

	for i := 0; i < resolveAttempts; i++ {
		log.Info().Msg("validating record")

		_, err := net.LookupHost(*args.Name)
		if err != nil {
			dnsErr, ok := err.(*net.DNSError)
			if !ok || !dnsErr.IsNotFound {
				l.Err(err).Msg("LookupHost failed")
				return
			}
		}

		time.Sleep(time.Minute * 1)
	}

	l.Info().Msg("done")
}

func start(ctx context.Context, yc *ycsdk.SDK, r53 *r53sdk.Route53) error {
	req := instancegroup.ListInstanceGroupsRequest{
		FolderId: *args.FolderID,
		Filter:   fmt.Sprintf("name=\"%s\"", *args.Group),
		PageSize: 1,
	}

	l, err := yc.InstanceGroup().InstanceGroup().List(ctx, &req)
	if err != nil {
		return fmt.Errorf("unable to get instance group id: %w", err)
	}

	groupID := l.InstanceGroups[0].Id

	log.Info().Str("groupId", groupID)

	t := time.NewTicker(time.Duration(*args.Timeout) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("terminating igdns")
			return nil
		case <-t.C:
			refresh(ctx, yc, r53, groupID)
		}
	}
}
