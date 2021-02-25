package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type state struct {
	cds    *template.Template
	api    *ycsdk.SDK
	ctx    context.Context
	cfg    map[string]string
	groups map[string][]string
}

func setup() error {
	if err := parser.Parse(os.Args); err != nil {
		return err
	}

	log.Logger = getlog(*args.LogLevel)

	return nil
}

func readcfg(file string) (map[string]string, error) {
	bin, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	cfg := map[string]string{}

	lnum := 0

	lines := bufio.NewScanner(bin)
	for lines.Scan() {
		lnum++

		kv := strings.SplitN(lines.Text(), ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid Name to ID mapping at line %d", lnum)
		}
		cfg[kv[0]] = kv[1]
	}

	return cfg, nil
}

func equals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func change(tempdir, filename string, content []byte) error {
	l := log.With().Str("filename", filename).Logger()

	tmp, err := ioutil.TempFile(tempdir, "cds")
	if err != nil {
		return fmt.Errorf("unable to create temp file: %w", err)
	}

	l = l.With().Str("tempfile", tmp.Name()).Logger()

	if _, err := tmp.Write(content); err != nil {
		return fmt.Errorf("unable to write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("unable to close temp file: %w", err)
	}

	if err := os.Rename(tmp.Name(), filename); err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	return nil
}

func refresh(st *state) {
	log.Info().Msg("refreshing CDS configuration")

	groups := map[string][]string{}
	for name, id := range st.cfg {
		l := log.With().Str("name", name).Str("id", id).Logger()
		l.Info().Msg("processing group")

		hosts, err := ycListInstances(st.ctx, st.api, id)
		if err != nil {
			l.Err(err).
				Msg("skipping refresh because unable to list instances")
			return
		}

		for i, h := range hosts {
			if h == "" {
				l.Info().
					Int("index", i).
					Msg("skipping refresh because the host isn't initialized")
				return
			}
		}

		groups[name] = hosts
	}

	changed := false
	for name, hosts := range groups {
		if !equals(hosts, st.groups[name]) {
			log.Info().Str("group", name).Msg("detected change")

			changed = true
			break
		}
	}

	if !changed {
		log.Info().Msg("skipping refresh because no groups was changed")
		return
	}

	buf := &bytes.Buffer{}

	if err := st.cds.Execute(buf, groups); err != nil {
		log.Err(err).Msg("unable to render template")
		return
	}

	if err := change(".", *args.Output, buf.Bytes()); err != nil {
		log.Err(err).Msg("unable to update hosts configuration")
		return
	}

	log.Info().Msg("CDS configuration changed")
	st.groups = groups
}

func run() error {
	cds, err := template.ParseFiles(*args.Template)
	if err != nil {
		return fmt.Errorf("unable to parse CDS template: %w", err)
	}

	cfg, err := readcfg(*args.Mapping)
	if err != nil {
		return fmt.Errorf("unable to read mapping file: %w", err)
	}

	ctx := context.Background()

	api, err := ycConnect(ctx, *args.Token)
	if err != nil {
		return fmt.Errorf("unable to connect Yandex Cloud API: %w", err)
	}

	groups := map[string][]string{}
	for k := range cfg {
		groups[k] = []string{}
	}

	st := state{
		cds:    cds,
		api:    api,
		cfg:    cfg,
		ctx:    ctx,
		groups: groups,
	}

	t := time.NewTicker(time.Duration(*args.Timeout) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("terminating cds")
			return nil
		case <-t.C:
			refresh(&st)
		}
	}
}

func main() {
	if err := setup(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := run(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
