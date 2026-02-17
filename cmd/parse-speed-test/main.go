package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	fap "github.com/hessu/go-aprs-fap"
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write CPU profile to file")
	filterError := flag.String("e", "", "print packets having the specified error code")
	flag.StringVar(filterError, "error", "", "print packets having the specified error code")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var ok, unsupported, fail int
	errCounts := make(map[string]int)

	start := time.Now()

	for scanner.Scan() {
		line := scanner.Text()

		// Strip APRS-IS unix timestamp prefix (first space-separated field)
		idx := strings.IndexByte(line, ' ')
		if idx < 0 {
			fail++
			errCounts["no space in line"]++
			continue
		}
		packet := line[idx+1:]

		// Skip comment lines
		if strings.HasPrefix(packet, "#") {
			continue
		}

		_, err := fap.Parse(packet)
		if err != nil {
			if errors.Is(err, fap.ErrTypeNotSupported) {
				unsupported++
			} else {
				fail++
			}
			var pe *fap.ParseError
			if errors.As(err, &pe) {
				errCounts[pe.Code]++
				if *filterError != "" && pe.Code == *filterError {
					fmt.Printf("%s [%s]\n", packet, err)
				}
			} else {
				errCounts[err.Error()]++
			}
		} else {
			ok++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	total := ok + unsupported + fail
	secs := elapsed.Seconds()
	rate := float64(total) / secs

	fmt.Printf("Parsed %d packets in %.3f seconds (%.0f packets/sec)\n", total, secs, rate)
	fmt.Printf("  OK: %d (%d unsupported), Failed: %d\n", ok+unsupported, unsupported, fail)

	if len(errCounts) > 0 {
		type errEntry struct {
			msg   string
			count int
		}
		entries := make([]errEntry, 0, len(errCounts))
		for msg, count := range errCounts {
			entries = append(entries, errEntry{msg, count})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].count > entries[j].count
		})

		fmt.Printf("\nError summary (%d unique errors):\n", len(errCounts))
		for _, e := range entries {
			fmt.Printf("  %6d  %s\n", e.count, e.msg)
		}
	}
}
