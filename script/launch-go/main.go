package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/mem"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var wg sync.WaitGroup

func ptyShell(command string) error {
	// Create arbitrary command.
	c := exec.Command(command)

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.Signal(30))
	go func() {
		for range ch {
			if err = pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Error().Msgf("failed to resize pty: %v", err)
			}
		}
	}()
	ch <- syscall.Signal(30)                      // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	return nil
}
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	memSettingFile := "/app/memorysize.txt"
	jvmArgsFile := "/app/userjvmargs.txt"
	_, memSettingStat := os.Stat(memSettingFile)
	_, jvmArgsStat := os.Stat(jvmArgsFile)
	var memSetting string
	var jvmArgs string
	vmi, _ := mem.VirtualMemory()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.IsNotExist(memSettingStat) || os.IsNotExist(jvmArgsStat) {
		log.Warn().Msgf("Memory size or JVM args file not found, using default values.")
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		log.Warn().Msgf("Mem args: %s", memSetting)
		jvmArgs = "-XX:UseZGC -XX:ZGCenerational"
		log.Warn().Msgf("JVM args: %s", jvmArgs)
	}
	memSettingByte, err := os.ReadFile(memSettingFile)
	if err != nil {
		log.Warn().Msgf("Memory size file read error, using default values.")
		vmi, _ = mem.VirtualMemory()
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		log.Warn().Msgf("Mem args: %s", memSetting)
		jvmArgs = "-XX:UseZGC -XX:ZGCener"
		log.Warn().Msgf("JVM args: %s", jvmArgs)
	}
	jvmArgsByte, err := os.ReadFile(jvmArgsFile)
	if err != nil {
		log.Warn().Msgf("Memory size file read error, using default values.")
		vmi, _ = mem.VirtualMemory()
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		log.Warn().Msgf("Mem args: %s", memSetting)
		jvmArgs = "-XX:UseZGC" + " " + "-XX:ZGCener"
		log.Warn().Msgf("JVM args: %s", jvmArgs)
	}
	memSetting = string(memSettingByte)
	jvmArgs = string(jvmArgsByte)
	app, _ := os.ReadDir("/app")
	appStat, _ := os.Stat("/app")
	if !appStat.IsDir() {
		log.Fatal().Msgf("this image need dir to store server data, not is a file")
	}
	//if reflect.ValueOf(appStat.Sys()).Elem().FieldByName("perm").Field(0).Int() < 0400 {
	//
	//}
	for _, path := range app {
		err = os.Link("/app"+path.Name(), "/jbin"+path.Name())
		if err != nil {
			log.Error().AnErr("Fill to link file", err).Msgf("file name: %s", path.Name())
		}
	}
	log.Info().Msgf("server launching...")
	log.Info().Msgf("this server have %d bytes memory can use", int(vmi.Available))
	launchCmd := "java" + " " + "-jar" + " " + "/jbin/server.jar" + " " + memSetting + " " + jvmArgs
	log.Printf("launch command: %s", launchCmd)
	_ = os.Chdir("/jbin")
	err = ptyShell(launchCmd)
	if err != nil {
		log.Fatal().AnErr("Fill to launch server", err).Msg("")
	}
	if len(app) > 2 {
		os.Exit(0)
	}
	log.Info().Msgf("this server is first launch, no jar file found in /app, will move files to /app")
	source, _ := os.ReadDir("/jbin")
	//ctx := context.Background()
	for _, path := range source {
		switch path.Name() {
		case "server.jar":
			continue
		default:
			wg.Add(1)
			go func() {
				err = os.Rename("/jbin"+path.Name(), "/app"+path.Name())
				if err != nil {
					log.Error().AnErr("Fill to move file", err).Msgf("file name: %s", path.Name())
				}
			}()
			wg.Done()
		}
	}
	wg.Wait()
	return
}
