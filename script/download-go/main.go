package main

import (
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/get"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/info"
	"github.com/Mr-Ao-Dragon/MCSL-Sync-Golang-SDK/setup"
	"log"
	"os"
	"sort"
	"strconv"
)

func main() {
	client := setup.InitSetupData(
		"sync.mcsl.com.cn",
		true,
		"",
		"mohist",
		os.Getenv("MCVersion"),
		"",
		"/app",
	)
	data := new(info.CoreInfo)
	data.GetCoreBuildListSingleMCVersion(*client)
	strKey := make([]int, 0)
	for k := range data.HistoryVersion {
		numKey, err := strconv.Atoi(k)
		if err != nil {
			log.Panicf("无法对列表进行转换")
		}
		strKey = append(strKey, numKey)
	}
	sort.Ints(strKey)
	err := get.Download(*client, data.HistoryVersion[strconv.Itoa(strKey[len(strKey)-1])], "server.jar")
	if err != nil {
		log.Fatalf("fail: %v", err)
	}

}
