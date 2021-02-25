package main

import (
	"github.com/rs/zerolog"
	"github.com/vbogretsov/argparse"
)

const usage = `
igdns

Assigns a Route53 domain name to the Yandex Cloud instance group proivded.
Once the group is being scaled a DNS record will be updated.

If the process finds $(pwd)/.env it will try to set environment variables from it.
The file lines format should be KEY=VALUE.

`

var parser = argparse.NewParser("igdns", usage)

type argT struct {
	Token    *string
	FolderID *string
	Group    *string
	Zone     *string
	Name     *string
	TTL      *int
	Comment  *string
	Timeout  *int
	LogLevel *string
}

var args = argT{
	Token: parser.String("k", "token", &argparse.Options{
		Default: "",
		Env:     argparse.Env{Name: "YANDEX_TOKEN"},
		Help:    "Yandex IAM token.",
	}),
	FolderID: parser.String("f", "folder", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGDNS_FOLDER"},
		Help:     "Instance Group parent Folder ID.",
	}),
	Group: parser.String("g", "group", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGDNS_GROUP"},
		Help:     "Instance Group name.",
	}),
	Zone: parser.String("z", "zone", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGDNS_ZONE"},
		Help:     "Route53 Public Hosted Zone ID.",
	}),
	Name: parser.String("n", "name", &argparse.Options{
		Required: true,
		Env:      argparse.Env{Name: "IGDNS_NAME"},
		Help:     "",
	}),
	TTL: parser.Int("", "ttl", &argparse.Options{
		Default: 300,
		Env:     argparse.Env{Name: "IGDNS_TTL"},
		Help:    "TTL for the DNS record being updated.",
	}),
	Comment: parser.String("c", "comment", &argparse.Options{
		Default: "Managed by igdns",
		Env:     argparse.Env{Name: "IGDNS_COMMENT"},
		Help:    "Comment for the DNS records being updated.",
	}),
	LogLevel: parser.Selector("l", "loglevel", loglevels, &argparse.Options{
		Default: zerolog.InfoLevel.String(),
		Env:     argparse.Env{Name: "IGDNS_LOG_LEVEL"},
		Help:    "Log level",
	}),
	Timeout: parser.Int("t", "timeout", &argparse.Options{
		Default: 5,
		Env:     argparse.Env{Name: "IGDNS_TIMEOUT"},
		Help:    "DNS refresh timeout.",
	}),
}
