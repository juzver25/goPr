package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	url             = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval    = 5 * time.Second
	httpTimeout     = 3 * time.Second
	maxErrorReports = 3
)

func main() {
	client := &http.Client{Timeout: httpTimeout}
	errCount := 0

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	for {
		if ok := pollOnce(client); !ok {
			errCount++
		} else {
			errCount = 0
		}
		if errCount >= maxErrorReports {
			fmt.Println("Unable to fetch server statistic.")

		}
		<-t.C
	}
}

func pollOnce(client *http.Client) bool {
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	br := bufio.NewScanner(resp.Body)
	if !br.Scan() {
		return false
	}
	line := strings.TrimSpace(br.Text())
	parts := strings.Split(line, ",")
	if len(parts) != 7 {
		return false
	}

	nums := make([]float64, 0, 7)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return false
		}
		nums = append(nums, v)
	}

	load := nums[0]

	memTotal := nums[1]
	memUsed := nums[2]

	diskTotal := nums[3]
	diskUsed := nums[4]

	netBW := nums[5]
	netUsed := nums[6]

	if load > 30 {

		fmt.Printf("Load Average is too high: %g\n", load)
	}

	if memTotal > 0 {
		memPct := (memUsed / memTotal) * 100
		if memPct > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memPct)
		}
	}

	if diskTotal > 0 {
		usedPct := (diskUsed / diskTotal) * 100
		if usedPct > 90 {
			freeBytes := diskTotal - diskUsed
			freeMB := freeBytes / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeMB)
		}
	}

	if netBW > 0 {
		usedPct := (netUsed / netBW) * 100
		if usedPct > 90 {
			freeBytesPerSec := netBW - netUsed
			freeMbit := (freeBytesPerSec * 8) / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}
	}

	return true
}
