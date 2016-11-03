// Package main reads lines from files and writes them to a Windows
// event log.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andrewkroh/sys/windows/svc/eventlog"
	"github.com/juju/ratelimit"
)

// EventCreate.exe has valid event IDs in the range of 1-1000 where each
// event message requires a single parameter.
const eventCreateMsgFile = "%SystemRoot%\\System32\\EventCreate.exe"

var (
	file     = flag.String("f", "", "file to read")
	log      = flag.String("l", "EventSystem", "event source name")
	id       = flag.Uint("id", 512, "event id")
	max      = flag.Uint("max", 0, "maximum events to write")
	rate     = flag.Float64("rate", 0, "rate limit in events per second")
	interval = flag.Duration("interval", 0, "interval at which to send events")

	install  = flag.Bool("install", false, "install new event source")
	provider = flag.String("provider", "Application", "provider name to install")
	source   = flag.String("source", "", "source name to install, must be specified to install new event source")
)

type eventgen struct {
	file  *os.File          // Input file to read from.
	log   *eventlog.Log     // Windows event log.
	tb    *ratelimit.Bucket // Token bucker for rate limiting.
	count uint32            // Total number of events written.
	max   uint32            // Max lines to read.
	done  chan struct{}     // Channel used to signal to stop.
}

func (eg *eventgen) installSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		once := sync.Once{}
		for _ = range c {
			once.Do(func() { close(eg.done) })
		}
	}()
}

func (eg *eventgen) reportEvents(wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(eg.file)
	for scanner.Scan() {
		if eg.tb != nil {
			for !eg.tb.WaitMaxDuration(1, time.Second) {
				select {
				case <-eg.done:
					return
				default:
				}
			}
		}

		eg.log.Report(eventlog.Info, uint32(*id), []string{scanner.Text()})

		numRead := atomic.AddUint32(&eg.count, 1)
		if eg.max != 0 && numRead >= eg.max {
			return
		}

		select {
		case <-eg.done:
			return
		default:
		}
	}
}

func registerSource(provider, sourceName string) error {
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if sourceName == "" {
		return fmt.Errorf("source cannot be empty")
	}

	alreadyInstalled, err := eventlog.InstallAsEventCreate(provider, sourceName,
		eventlog.Error|eventlog.Warning|eventlog.Info|eventlog.Success|
			eventlog.AuditSuccess|eventlog.AuditFailure)
	if err != nil {
		return err
	}

	if alreadyInstalled {
		fmt.Printf("%s/%s already exists\n", provider, sourceName)
		return nil
	}

	fmt.Printf("%s/%s installed\n", provider, sourceName)
	return nil
}

func main() {
	flag.Parse()

	if *install {
		err := registerSource(*provider, *source)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if *file == "" {
		fmt.Fprintln(os.Stderr, "-f is required")
		flag.Usage()
		os.Exit(1)
	}

	// Open a handle to the event log.
	log, err := eventlog.Open(*log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening eventlog:", err)
		os.Exit(1)
	}

	// Open the file on disk.
	file, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	eg := &eventgen{
		file: file,
		max:  uint32(*max),
		done: make(chan struct{}),
		log:  log,
	}
	eg.installSignalHandler()

	// Rate limit writing using a token bucket.
	if *rate != 0 {
		eg.tb = ratelimit.NewBucketWithRate(*rate, int64(math.Ceil(*rate)))
		eg.tb.TakeAvailable(eg.tb.Available())
	} else if *interval != 0 {
		eg.tb = ratelimit.NewBucket(*interval, 1)
		eg.tb.TakeAvailable(eg.tb.Available())
	}

	start := time.Now()

	// Start a new worker to read lines from the file.
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go eg.reportEvents(wg)
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("elapsed time:", elapsed)
	fmt.Println("event count:", atomic.LoadUint32(&eg.count))
}
