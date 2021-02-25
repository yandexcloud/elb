package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	r53sdk "github.com/aws/aws-sdk-go/service/route53"
	"github.com/rs/zerolog/log"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const envFile = ".env"

func r53Connect() (*r53sdk.Route53, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	cred := stscreds.NewCredentials(sess, os.Getenv("AWS_ROLE"))
	api := r53sdk.New(sess, &aws.Config{Credentials: cred})

	return api, nil
}

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

func setenv(file string) error {
	env, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to open env file: %w", err)
	}

	lnum := 0

	lines := bufio.NewScanner(env)
	for lines.Scan() {
		lnum++

		kv := strings.SplitN(lines.Text(), "=", 2)
		if err := os.Setenv(kv[0], kv[1]); err != nil {
			fmt.Printf("[WARN] unable to set env %s=%s\n", kv[0], kv[1])
		}
	}

	return nil
}

func setup() error {
	if err := setenv(envFile); err != nil {
		return err
	}

	if err := parser.Parse(os.Args); err != nil {
		return err
	}

	log.Logger = getlog(*args.LogLevel)

	return nil
}

func run() error {
	log.Debug().
		Str("group", *args.Group).
		Str("zone", *args.Zone).
		Str("name", *args.Name).
		Int("timeout", *args.Timeout).
		Msg("args")

	ctx := context.Background()

	yc, err := ycConnect(ctx, *args.Token)
	if err != nil {
		return fmt.Errorf("unable to initialize Yandex Cloud SDK: %w", err)
	}

	r53, err := r53Connect()
	if err != nil {
		return fmt.Errorf("unable to initialize S3 SDK: %w", err)
	}

	return start(ctx, yc, r53)
}

func main() {
	if err := setup(); err != nil {
		fmt.Printf("[FATAL] %v\n", err)
		os.Exit(1)
	}

	if err := run(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
