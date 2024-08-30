package main

import (
	"github.com/shirou/gopsutil/v4/mem"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

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
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
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
	memSettingFile := "/app/memorysize.txt"
	jvmArgsFile := "/app/userjvmargs.txt"
	_, memSettingStat := os.Stat(memSettingFile)
	_, jvmArgsStat := os.Stat(jvmArgsFile)
	var memSetting string
	var jvmArgs string
	if os.IsNotExist(memSettingStat) || os.IsNotExist(jvmArgsStat) {
		log.Println("Memory size or JVM args file not found, using default values.")
		vmi, _ := mem.VirtualMemory()
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		jvmArgs = "-XX:UseZGC -XX:ZGCenerational"
	}
	memSettingByte, err := os.ReadFile(memSettingFile)
	if err != nil {
		log.Println("Memory size file read error, using default values.")
		vmi, _ := mem.VirtualMemory()
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		jvmArgs = "-XX:UseZGC -XX:ZGCener"
	}
	jvmArgsByte, err := os.ReadFile(jvmArgsFile)
	if err != nil {
		log.Println("Memory size file read error, using default values.")
		vmi, _ := mem.VirtualMemory()
		memSetting = "-Xmx" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 3))), 10) + "G" + "-Xms" + strconv.FormatUint(vmi.Available/uint64(math.Trunc(math.Pow(1024, 5))), 10)
		jvmArgs = "-XX:UseZGC -XX:ZGCener"
	}
	memSetting = string(memSettingByte)
	jvmArgs = string(jvmArgsByte)
	err = ptyShell("java " + memSetting + " " + jvmArgs + " -jar /app/server.jar")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
