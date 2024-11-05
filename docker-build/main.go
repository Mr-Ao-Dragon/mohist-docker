package main

import (
	"context"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/info"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/setup"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
)

var wg sync.WaitGroup

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	MCSLSyncClient := setup.InitSetupData(
		"sync.mcsl.com.cn",
		true,
		"",
		"Mohist")
	coreData := new(info.CoreInfo)
	errs := coreData.GetCoreSupportMcList(*MCSLSyncClient)
	for _, err := range errs {
		if err != nil {
			log.Error().AnErr("GetCoreSupportMcList", err).Msg("")
		}
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal().AnErr("NewClientWithOpts", err).Msg("init docker sdk failed")
	}
	defer dockerClient.Close()
	buildContext, err := os.Open(os.Getenv("GITHUB_WORKSPACE"))
	if err != nil {
		log.Fatal().AnErr("Open", err).Msg("open workspace failed")
	}
	defer buildContext.Close()
	for _, version := range coreData.SupportMcVersion {
		err = os.Setenv("MC_VERSION", version)
		if err != nil {
			log.Error().AnErr("Setenv", err).Msgf("Setenv MC_VERSION = %s failed", version)
			continue
		}
		go func() {
			buildOpts := types.ImageBuildOptions{
				Context:     buildContext,
				Tags:        []string{"whf-studio/mohist-docker" + ":" + "build" + "-" + version},
				Dockerfile:  os.Getenv("GITHUB_WORKSPACE" + "/" + "preBuild.Dockerfile"),
				NoCache:     true,
				Squash:      true,
				ForceRemove: true,
				PullParent:  true,
				Platform:    "",
			}
			wg.Add(1)
			log.Info().Msgf("Building image...")
			respBody, err := dockerClient.ImageBuild(context.Background(), buildContext, buildOpts)
			if err != nil {
				log.Error().AnErr("ImageBuild", err).Msgf("build tag: %s failed", buildOpts.Tags)
			}
			log.Info().Msgf("Building image finished")
			log.Info().Msgf("os type: %s", respBody.OSType)
			wg.Done()
		}()
	}
	wg.Wait()
}
