package integration

import (
	"context"
	"fmt"
	devtools "github.com/elastic/elastic-agent/dev-tools/mage"
	"github.com/elastic/elastic-agent/dev-tools/mage/stack"
	"github.com/elastic/elastic-agent/internal/pkg/release"
	"github.com/magefile/mage/sh"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	buildDir = "testing/hello"

	metaDir           = "../../_meta"
	snapshotEnv       = "SNAPSHOT"
	devEnv            = "DEV"
	externalArtifacts = "EXTERNAL"
	configFile        = "elastic-agent.yml"
)

// Up teardown docker environment
func Up(ctx context.Context, version string, dockerImage string) error {
	fmt.Println("Load elastic agent image")
	directory, _ := filepath.Abs(filepath.Join("build/distributions", dockerImage))
	fmt.Println(directory)

	fmt.Println("Tag elastic agent image")
	version = version + "-SNAPSHOT"
	// we need to tag the loaded image because its tag relates to the target branch

	fmt.Println("Deploy stack")
	service := &stack.Service{
		Context: context.Background(),
	}

	//profile := deploy.NewServiceRequest("fleet")
	env := map[string]string{}
	env["ELASTIC_AGENT_IMAGE_REF_OVERRIDE"] = "docker.elastic.co/observability-ci/elastic-agent-complete:8.3.0-SNAPSHOT-amd64"
	//config.Init()
	err := service.Start(ctx, env, func() error {
		fmt.Println("stack has been deployed")
		return nil
	})
	fmt.Println("stack has been deployed", err)
	return err
}

// Down teardown docker environment
func Down(ctx context.Context) error {
	service := &stack.Service{
		Context: context.Background(),
	}
	env := map[string]string{}
	err := service.Stop(ctx, env)
	return err
}

func BuildBinary() error {
	start := time.Now()
	defer func() { fmt.Println("build binary ran for", time.Since(start)) }()
	Mkdir("testing/hello")
	err := GenerateConfig()
	if err != nil {
		return err
	}
	err = RunGo("version")
	if err != nil {
		return err
	}
	err = RunGo("env")
	if err != nil {
		return err
	}
	buildArgs := devtools.DefaultBuildArgs()
	buildArgs.OutputDir = "testing/hello"
	injectBuildVars(buildArgs.Vars)

	return devtools.Build(buildArgs)
}

// Clean up dev environment.
func Clean() {
	os.RemoveAll("testing/hello")
}

func LoadImage() error {
	start := time.Now()
	defer func() { fmt.Println("stack up ran for", time.Since(start)) }()
	//	ctx := context.Background()
	fmt.Println("starting stack up")
	version, found := os.LookupEnv("BEAT_VERSION")
	if !found {
		version = release.Version()
	}
	fmt.Println("Found version ", version)
	fmt.Println("Set default env variables")
	os.Setenv("PLATFORMS", "linux/amd64")
	devtools.Platforms = devtools.NewPlatformList("linux/amd64")
	os.Setenv("PACKAGES", "DOCKER")
	devtools.PACKAGES = "DOCKER"
	os.Setenv(snapshotEnv, "true")
	devtools.Snapshot = true
	os.Setenv(externalArtifacts, "true")
	devtools.ExternalBuild = true
	os.Setenv(devEnv, "true")
	devtools.DevBuild = true

	// produce docker package
	//packageAgent([]string{
	//		"linux-x86_64.tar.gz",
	//	}, devtools.UseElasticAgentDockerTestPackaging, true)

	//	fmt.Println("docker image created")
	//	stack.Up(ctx, version, "elastic-agent-complete-8.3.0-SNAPSHOT-linux-amd64.docker.tar.gz")
	return nil
}

func buildVars() map[string]string {
	vars := make(map[string]string)
	vars["github.com/elastic/elastic-agent/internal/pkg/release.snapshot"] = "true"

	if isDevFlag, devFound := os.LookupEnv("DEV"); devFound {
		if isDev, err := strconv.ParseBool(isDevFlag); err == nil && isDev {
			vars["github.com/elastic/elastic-agent/internal/pkg/release.allowEmptyPgp"] = "true"
			vars["github.com/elastic/elastic-agent/internal/pkg/release.allowUpgrade"] = "true"
		}
	}

	return vars
}

func injectBuildVars(m map[string]string) {
	for k, v := range buildVars() {
		m[k] = v
	}
}

// Mkdir returns a function that create a directory.
func Mkdir(dir string) func() error {
	return func() error {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory: %v, error: %+v", dir, err)
		}
		return nil
	}
}
func RunGo(args ...string) error {
	return sh.RunV("go", args...)
}
func GenerateConfig() error {
	return sh.Copy(filepath.Join(buildDir, configFile), filepath.Join(metaDir, configFile))
}
