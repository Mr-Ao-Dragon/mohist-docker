package main

import (
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/get"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/info"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/setup"
	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
	"os"
	"sort"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	client := setup.InitSetupData(
		"sync.mcsl.com.cn",
		true,
		"",
		"Mohist",
		os.Getenv("MCVersion"),
		"",
		"/jbin",
	)
	data := new(info.CoreInfo)
	data.GetCoreBuildListSingleMCVersion(*client)
	strKey := make([]string, 0)
	for k := range data.HistoryVersion {
		strKey = append(strKey, k)
	}
	sort.Strings(strKey)
	_ = os.Chdir("/")
	_ = os.Mkdir("jbin", 0644)
	_ = os.Chdir("/jbin")
	errs := get.Download(*client, data.HistoryVersion[strKey[len(strKey)-1]], "server.jar")
	if errs != nil {
		log.Fatal().AnErr("fail to downlod", errs).Msg("")
	}
}
