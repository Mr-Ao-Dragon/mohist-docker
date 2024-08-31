package main

import (
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/get"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/info"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/setup"
	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"
	"os"
	"sort"
	"strconv"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	client := setup.InitSetupData(
		"sync.mcsl.com.cn",
		true,
		"",
		"mohist",
		os.Getenv("MCVersion"),
		"",
		"/jbin",
	)
	data := new(info.CoreInfo)
	data.GetCoreBuildListSingleMCVersion(*client)
	strKey := make([]int, 0)
	for k := range data.HistoryVersion {
		numKey, err := strconv.Atoi(k)
		if err != nil {
			log.Panic().AnErr("key", err).Msg("无法对列表进行转换")
		}
		strKey = append(strKey, numKey)
	}
	sort.Ints(strKey)
	os.Chdir("/")
	os.Mkdir("jbin", 0644)
	os.Chdir("/jbin")
	err := get.Download(*client, data.HistoryVersion[strconv.Itoa(strKey[len(strKey)-1])], "server.jar")
	if err != nil {
		log.Fatal().AnErr("fail to downlod", err)
	}
}
