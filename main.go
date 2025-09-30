package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type HistogramBucket struct {
	DistinctRange float64 `json:"distinct_range"`
	NumEq         int     `json:"num_eq"`
	NumRange      int     `json:"num_range"`
	UpperBound    string  `json:"upper_bound"`
}

type Statistics struct {
	Columns      []string          `json:"columns"`
	HistoBuckets []HistogramBucket `json:"histo_buckets"`
}

const (
	binaryName = "cockroach"
	intCol     = "i"
	stringCol  = "s"
)

type span struct {
	lo, hi int
}

func (s span) String() string {
	return fmt.Sprintf("[%d-%d]", s.lo, s.hi)
}

type spans []span

func (s spans) String() string {
	var sb strings.Builder
	for i, sp := range s {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sp.String())
	}
	return sb.String()
}

var (
	valRanges = spans{{1, 10_000}, {100_000, 110_000}}
)

func main() {
	// Read SQL file.
	sqlFile, err := os.Open("make_histograms.sql")
	if err != nil {
		log.Fatal("Failed to open make_histograms.sql:", err)
	}
	defer sqlFile.Close()

	sqlContent, err := io.ReadAll(sqlFile)
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Run cockroach demo with the SQL file.
	fmt.Println("Starting CockroachDB demo and executing SQL...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryName, "demo", "--execute", string(sqlContent))
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Fatal("Command failed with stderr:", string(exitErr.Stderr))
		}
		log.Fatal("Failed to execute cockroach demo:", err)
	}

	// Parse output to extract JSON from the last query.
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	var jsonResult string
	// Look for the line starting with quoted JSON.
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, `"[`) && strings.HasSuffix(line, `]"`) {
			// Remove the surrounding quotes.
			jsonResult = line[1 : len(line)-1]
			// Unescape the JSON.
			jsonResult = strings.ReplaceAll(jsonResult, `""`, `"`)
			break
		}
	}

	if jsonResult == "" {
		log.Fatal("Could not find JSON result in output")
	}

	// Parse the JSON result.
	var statistics []Statistics
	err = json.Unmarshal([]byte(jsonResult), &statistics)
	if err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	// Find histograms for columns "i" and "s".
	var iHistogram, sHistogram *Statistics
	for i := range statistics {
		if len(statistics[i].Columns) == 1 {
			if statistics[i].Columns[0] == intCol {
				iHistogram = &statistics[i]
			} else if statistics[i].Columns[0] == stringCol {
				sHistogram = &statistics[i]
			}
		}
	}

	if iHistogram == nil {
		log.Fatalf("Could not find histogram for column %q", intCol)
	}
	if sHistogram == nil {
		log.Fatalf("Could not find histogram for column %q", stringCol)
	}

	// Count matches for column int histogram.
	intCount := 0
	strCount := 0
	total := 0
	for _, sp := range valRanges {
		total += sp.hi - sp.lo + 1
		for v := sp.lo; v <= sp.hi; v++ {
			if nonEmptyIntBucket(iHistogram.HistoBuckets, v) {
				intCount++
			}
			if nonEmptyStringBucket(sHistogram.HistoBuckets, strconv.Itoa(v)) {
				strCount++
			}
		}
	}

	// Output results.
	fmt.Printf("Column 'i' histogram matches %d/%d values in the ranges %s.\n", intCount, total, valRanges)
	fmt.Printf("Column 's' histogram matches %d/%d values in the ranges %s.\n", strCount, total, valRanges)
}

func nonEmptyIntBucket(buckets []HistogramBucket, val int) bool {
	var prevUpperBound int
	for i, bucket := range buckets {
		currUpperBound, err := strconv.Atoi(bucket.UpperBound)
		if err != nil {
			panic(fmt.Sprintf("could not parse upper bound %q as int: %v", bucket.UpperBound, err))
		}
		// First, check for exact upper_bound match.
		if currUpperBound == val {
			return true
		}
		// Next, check for a range match.
		if val < currUpperBound && (i == 0 || val > prevUpperBound) {
			return bucket.NumRange > 0
		}
		prevUpperBound = currUpperBound
	}
	return false
}

func nonEmptyStringBucket(buckets []HistogramBucket, val string) bool {
	var prevUpperBound string
	for i, bucket := range buckets {
		currUpperBound := bucket.UpperBound
		// First, check for exact upper_bound match.
		if currUpperBound == val {
			return true
		}
		// Next, check for a range match.
		if val < currUpperBound && (i == 0 || val > prevUpperBound) {
			return bucket.NumRange > 0
		}
		prevUpperBound = currUpperBound
	}
	return false
}
