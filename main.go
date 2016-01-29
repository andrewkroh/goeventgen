// Package main reads lines from files and writes them to a Windows
// event log.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/andrewkroh/sys/windows/svc/eventlog"
)

// EventCreate.exe has valid event IDs in the range of 1-1000 where each
// event message requires a single parameter.
const eventCreateMsgFile = "%SystemRoot%\\System32\\EventCreate.exe"

var (
	file = flag.String("f", "", "file to read")
	log  = flag.String("l", "Application", "event log")
)

func main() {
	flag.Parse()

	if *file == "" {
		fmt.Fprintln(os.Stderr, "-f is required")
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	log, err := eventlog.Open(*log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening eventlog:", err)
		os.Exit(1)
	}

	var count uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		log.Report(eventlog.Info, 1, []string{scanner.Text()})
		count++
	}
	fmt.Println("Count:", count)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading file:", err)
		os.Exit(1)
	}
}
