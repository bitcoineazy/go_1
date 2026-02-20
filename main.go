package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var errorCount int

	for {
		if !fetchStats(client, &errorCount) {
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func fetchStats(client *http.Client, errorCount *int) bool {
	resp, err := client.Get("http://srv.msk01.gigacorp.local/_stats")
	if err != nil {
		*errorCount++
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		*errorCount++
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		*errorCount++
		return false
	}

	data := strings.TrimSpace(string(body))
	parts := strings.Split(data, ",")
	if len(parts) != 7 {
		*errorCount++
		return false
	}

	values := make([]int64, 7)
	for i, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			*errorCount++
			return false
		}
		values[i] = v
	}

	// Reset error count on successful fetch
	*errorCount = 0

	loadAvg := values[0]
	memTotal := values[1]
	memUsed := values[2]
	diskTotal := values[3]
	diskUsed := values[4]
	netBandwidth := values[5]
	netUsed := values[6]

	// Check Load Average (threshold: > 30)
	if loadAvg > 30 {
		fmt.Printf("Load Average is too high: %d\n", loadAvg)
	}

	// Check Memory usage (threshold: > 80%)
	if memTotal > 0 {
		memPercent := memUsed * 100 / memTotal
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memPercent)
		}
	}

	// Check Disk space (threshold: > 90% used)
	if diskTotal > 0 {
		diskPercent := diskUsed * 100 / diskTotal
		if diskPercent > 90 {
			freeDisk := (diskTotal - diskUsed) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDisk)
		}
	}

	// Check Network bandwidth (threshold: > 90% used)
	if netBandwidth > 0 {
		netPercent := netUsed * 100 / netBandwidth
		if netPercent > 90 {
			freeNet := (netBandwidth - netUsed) / 1000000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeNet)
		}
	}

	return true
}
