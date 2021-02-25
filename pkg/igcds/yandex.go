package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func ycConnect(ctx context.Context, token string) (*ycsdk.SDK, error) {
	if token == "" {
		return ycsdk.Build(ctx, ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		})
	}
	return ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.NewIAMTokenCredentials(token),
	})
}

func ycListInstances(ctx context.Context, api *ycsdk.SDK, id string) ([]string, error) {
	req := instancegroup.ListInstanceGroupInstancesRequest{
		InstanceGroupId: id,
		// NOTE: we do not expect large instance groups
		PageSize: 1000,
	}

	res, err := api.InstanceGroup().InstanceGroup().ListInstances(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("unable to list instances: %w", err)
	}

	hosts := make([]string, len(res.Instances))
	for i, v := range res.Instances {
		hosts[i] = v.Fqdn
	}

	sort.Strings(hosts)
	return hosts, nil
}
