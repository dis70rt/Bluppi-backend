package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "sort"
    "sync"
    "sync/atomic"
    "syscall"
    "time"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/presences"
    "github.com/google/uuid"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
)

// ANSI Colors
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorBlue   = "\033[34m"
    ColorPurple = "\033[35m"
    ColorCyan   = "\033[36m"
    ColorGray   = "\033[37m"
)

type Stats struct {
    ActiveConnections int64
    RequestsTotal     int64
    ErrorsTotal       int64
    MessagesReceived  int64
    Latencies         []time.Duration
    mu                sync.Mutex
}

var globalStats = &Stats{}

func main() {
    target := flag.String("target", "localhost:50050", "The gateway address")
    users := flag.Int("users", 50, "Number of concurrent users")
    duration := flag.Duration("duration", 30*time.Second, "Test duration")
    rampUp := flag.Duration("ramp-up", 5*time.Second, "Ramp up time")
    flag.Parse()

    ctx, cancel := context.WithTimeout(context.Background(), *duration+*rampUp+2*time.Second)
    defer cancel()

    // Handle Ctrl+C
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        cancel()
    }()

    fmt.Printf("%sStarting Load Test on %s%s\n", ColorCyan, *target, ColorReset)
    fmt.Printf("Users: %d | Duration: %v | RampUp: %v\n\n", *users, *duration, *rampUp)

    var wg sync.WaitGroup
    startTime := time.Now()

    // UI Loop
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                printStats(startTime, *users)
            }
        }
    }()

    // Worker Pool
    limitr := time.NewTicker(*rampUp / time.Duration(*users))
    for i := 0; i < *users; i++ {
        <-limitr.C
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            worker(ctx, *target, id)
        }(i)
    }
    limitr.Stop()

    wg.Wait()
    printStats(startTime, *users)
    fmt.Println("\nLoad Test Completed.")
}

func worker(ctx context.Context, target string, id int) {
    // Connect
    conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        atomic.AddInt64(&globalStats.ErrorsTotal, 1)
        return
    }
    defer conn.Close()

    client := pb.NewPresenceGatewayClient(conn)

    // Mock User ID
    fakeUserID := uuid.New().String()
    md := metadata.New(map[string]string{
        "x-mock-user-id": fakeUserID,
    })
    outCtx := metadata.NewOutgoingContext(ctx, md)

    start := time.Now()

    // Subscribe
    stream, err := client.SubscribePresence(outCtx, &pb.SubscribeRequest{
        TargetUserIds: []string{fakeUserID}, // Subscribe to self to test loopback
    })

    if err != nil {
        atomic.AddInt64(&globalStats.ErrorsTotal, 1)
        return
    }

    latency := time.Since(start)
    
    globalStats.mu.Lock()
    globalStats.Latencies = append(globalStats.Latencies, latency)
    globalStats.mu.Unlock()

    atomic.AddInt64(&globalStats.RequestsTotal, 1)
    atomic.AddInt64(&globalStats.ActiveConnections, 1)
    defer atomic.AddInt64(&globalStats.ActiveConnections, -1)

    // Keep stream open
    for {
        select {
        case <-ctx.Done():
            return
        default:
            _, err := stream.Recv()
            if err != nil {
                // Don't count context cancellation as error
                if ctx.Err() == nil {
                    atomic.AddInt64(&globalStats.ErrorsTotal, 1)
                }
                return
            }
            atomic.AddInt64(&globalStats.MessagesReceived, 1)
        }
    }
}

func printStats(start time.Time, expectedUsers int) {
    globalStats.mu.Lock()
    latencies := make([]time.Duration, len(globalStats.Latencies))
    copy(latencies, globalStats.Latencies)
    globalStats.mu.Unlock()

    sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

    var p50, p90, p95, p99 time.Duration
    if len(latencies) > 0 {
        p50 = latencies[int(float64(len(latencies))*0.50)]
        p90 = latencies[int(float64(len(latencies))*0.90)]
        p95 = latencies[int(float64(len(latencies))*0.95)]
        p99 = latencies[int(float64(len(latencies))*0.99)]
    }

    active := atomic.LoadInt64(&globalStats.ActiveConnections)
    total := atomic.LoadInt64(&globalStats.RequestsTotal)
    errs := atomic.LoadInt64(&globalStats.ErrorsTotal)
    msgs := atomic.LoadInt64(&globalStats.MessagesReceived)
    elapsed := time.Since(start).Seconds()

	errColor := ColorGreen
    if errs > 0 {
        errColor = ColorRed
    }

    // Rate calculations
    rps := float64(total) / elapsed
    mps := float64(msgs) / elapsed

    // Clear screen/Line
    fmt.Print("\033[2J\033[H")

    fmt.Println("⚡ BLUPPI PRESENCE LOAD TEST")
    fmt.Println("===========================")
    fmt.Printf("Time Elapsed:  %s%.0fs%s\n", ColorBlue, elapsed, ColorReset)
    fmt.Printf("Active Users:  %s%d%s / %d\n", ColorGreen, active, ColorReset, expectedUsers)
    fmt.Printf("Total Requests: %d\n", total)
    fmt.Printf("Total Errors:   %s%d%s\n", errColor, errs, ColorReset)
    fmt.Printf("Messages Rx:    %d\n", msgs)
    
    fmt.Println("\n--- Throughput ---")
    fmt.Printf("Conn Rate:     %.2f req/s\n", rps)
    fmt.Printf("Msg Rate:      %.2f msg/s\n", mps)

    fmt.Println("\n--- Latency Distribution (Connect) ---")
    fmt.Printf("p50: %s%v%s\n", ColorCyan, p50, ColorReset)
    fmt.Printf("p90: %s%v%s\n", ColorBlue, p90, ColorReset)
    fmt.Printf("p95: %s%v%s\n", ColorYellow, p95, ColorReset)
    fmt.Printf("p99: %s%v%s\n", ColorRed, p99, ColorReset)
}