package main

import (
	"github.com/rs/zerolog"
	"github.com/vbogretsov/argparse"
)

const usage = `
igproxy

File based Envody CDS server. Pools information about an instance group's members
and updates the cluster configuration based on the Go template provided.

Example template:

resources:
- "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
  name: masters
  connect_timeout: 1s
  type: STRICT_DNS
  load_assignment:
    cluster_name: masters
    endpoints:
    - lb_endpoints:
    {{- range $host := .GroupName}}
      - endpoint:
          address:
            socket_address:
              address: {{ $host }}
              port_value: 80
    {{- end }}

The GroupName will be replaced to list of FQDNs of an instance group members.
The instance group should be specified in the configuration file like:
GroupName:GroupId

Example:

GroupName:cl18r909jrnv2l7p1ft3

Serveral groups can be provided.
`

var parser = argparse.NewParser("igcds", usage)

type argT struct {
	Token    *string
	Timeout  *int
	LogLevel *string
	Mapping  *string
	Template *string
	Output   *string
}

var args = argT{
	Token: parser.String("k", "token", &argparse.Options{
		Default: "",
		Env:     argparse.Env{Name: "YANDEX_TOKEN"},
		Help:    "Yandex IAM token.",
	}),
	LogLevel: parser.Selector("l", "loglevel", loglevels, &argparse.Options{
		Default: zerolog.InfoLevel.String(),
		Env:     argparse.Env{Name: "IGCDS_LOG_LEVEL"},
		Help:    "Log level.",
	}),
	Timeout: parser.Int("t", "timeout", &argparse.Options{
		Default: 5,
		Env:     argparse.Env{Name: "IGCDS_TIMEOUT"},
		Help:    "Refresh timeout.",
	}),
	Mapping: parser.String("m", "mapping", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGCDS_MAPPING"},
		Help:     "Path to mapping file.",
	}),
	Output: parser.String("o", "output", &argparse.Options{
		Default: "cds.yaml",
		Env:     argparse.Env{Name: "IGCDS_OUTPUT"},
		Help:    "Path to output CDS configuration.",
	}),
	Template: parser.String("f", "template", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGCDS_TEMPLATE"},
		Help:     "Path to cluster configuration template.",
	}),
}
