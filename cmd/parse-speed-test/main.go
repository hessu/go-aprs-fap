package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	fap "github.com/hessu/go-aprs-fap"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var ok, fail int

	start := time.Now()

	for scanner.Scan() {
		line := scanner.Text()

		// Strip APRS-IS unix timestamp prefix (first space-separated field)
		idx := strings.IndexByte(line, ' ')
		if idx < 0 {
			fail++
			continue
		}
		packet := line[idx+1:]

		_, err := fap.Parse(packet)
		if err != nil {
			fail++
		} else {
			ok++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	total := ok + fail
	secs := elapsed.Seconds()
	rate := float64(total) / secs

	fmt.Printf("Parsed %d packets in %.3f seconds (%.0f packets/sec)\n", total, secs, rate)
	fmt.Printf("  OK: %d  Failed: %d\n", ok, fail)
}
