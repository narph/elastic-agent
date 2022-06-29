// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package stack

import (
	"bytes"
	"context"
	"fmt"
	"go.elastic.co/apm"
	"os"
	"os/exec"
	"strings"
)

var elasticPackageBaseCommand = []string{"run", "github.com/elastic/elastic-package"}

type Service struct {
	Context context.Context
}

func (s *Service) Start(ctx context.Context, env map[string]string, waitCB func() error) error {
	services := "elasticsearch,fleet-server,kibana"

	version := "8.3.0"

	elasticPackageProfile := "default"
	args := append(elasticPackageBaseCommand, "stack", "up", "--daemon", "--verbose", "--version", version, "--services", services, "-p", elasticPackageProfile)

	span, _ := apm.StartSpanOptions(ctx, "Bootstrapping Elastic Package deployment", "elastic-package.manifest.bootstrap", apm.SpanOptions{
		Parent: apm.SpanFromContext(ctx).TraceContext(),
	})
	span.Context.SetLabel("args", args)
	span.Context.SetLabel("profile", "default")
	span.Context.SetLabel("services", services)
	span.Context.SetLabel("stackVersion", version)
	defer span.End()

	_, err := Execute(ctx, env, args...)
	return err
}

func (s *Service) Stop(ctx context.Context, env map[string]string) error {
	span, _ := apm.StartSpanOptions(ctx, "Destroying Elastic-Package deployment", "elastic-package.manifest.destroy", apm.SpanOptions{
		Parent: apm.SpanFromContext(ctx).TraceContext(),
	})
	span.Context.SetLabel("profile", "default")
	defer span.End()

	args := append(elasticPackageBaseCommand, "stack", "down", "--verbose")
	_, err := Execute(ctx, env, args...)
	return err
}

func Execute(ctx context.Context, env map[string]string, args ...string) (string, error) {
	span, _ := apm.StartSpanOptions(ctx, "Executing shell command", "shell.command.execute", apm.SpanOptions{
		Parent: apm.SpanFromContext(ctx).TraceContext(),
	})
	span.Context.SetLabel("workspace", ".")
	span.Context.SetLabel("command", "go")
	span.Context.SetLabel("arguments", args)
	span.Context.SetLabel("environment", env)
	defer span.End()

	cmd := exec.Command("go", args[0:]...)

	cmd.Dir = "."

	if len(env) > 0 {
		environment := os.Environ()

		for k, v := range env {
			environment = append(environment, fmt.Sprintf("%s=%s", k, v))
		}

		cmd.Env = environment
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	trimmedOutput := strings.Trim(out.String(), "\n")
	return trimmedOutput, nil
}
