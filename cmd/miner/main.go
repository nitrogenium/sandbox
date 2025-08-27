package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	pkgsolver "github.com/nitrogen/go-miner/pkg/solver"
	"github.com/nitrogen/go-miner/pkg/stratum"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MinerStats struct {
	StartTime      time.Time
	CyclesTotal    atomic.Uint64
	SolutionsTotal atomic.Uint64
	SharesAccepted atomic.Uint64
	SharesRejected atomic.Uint64
	LastCycles     atomic.Uint64
	LastSolutions  atomic.Uint64
	LastTime       time.Time
}

type Miner struct {
	// Configuration
	poolAddr string
	username string
	password string
	threads  int

	// Components
	client  *stratum.Client
	solvers []*pkgsolver.Solver
	logger  *zap.Logger

	// State
	currentWork *stratum.Work
	workMutex   sync.RWMutex
	extraNonce2 atomic.Uint64
	mining      atomic.Bool

	// Statistics
	stats MinerStats

	// Control
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewMiner(poolAddr, username, password string, threads int, logger *zap.Logger) *Miner {
	return &Miner{
		poolAddr: poolAddr,
		username: username,
		password: password,
		threads:  threads,
		logger:   logger,
		stopCh:   make(chan struct{}),
		stats: MinerStats{
			StartTime: time.Now(),
			LastTime:  time.Now(),
		},
	}
}

func (m *Miner) Start() error {
	m.logger.Info("Starting miner",
		zap.String("pool", m.poolAddr),
		zap.String("user", m.username),
		zap.Int("threads", m.threads))

	// Initialize solvers
	m.solvers = make([]*pkgsolver.Solver, m.threads)
	for i := 0; i < m.threads; i++ {
		m.solvers[i] = pkgsolver.NewSolver(1) // Each solver single-threaded
	}

	// Create Stratum client
	m.client = stratum.NewClient(m.poolAddr, m.username, m.password, m.logger)
	m.client.SetWorkHandler(m.handleNewWork)
	m.client.SetReconnectHandler(m.handleReconnect)

	// Connect to pool
	if err := m.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Start stats printer
	go m.printStats()

	return nil
}

func (m *Miner) Stop() {
	m.logger.Info("Stopping miner...")
	m.mining.Store(false)
	close(m.stopCh)
	m.client.Close()
	m.wg.Wait()
}

func (m *Miner) handleNewWork(work *stratum.Work) {
	m.logger.Info("New work received", zap.String("jobID", work.JobID))

	// Update current work
	m.workMutex.Lock()
	m.currentWork = work
	m.workMutex.Unlock()

	// Stop existing mining
	m.mining.Store(false)
	m.wg.Wait()

	// Start new mining
	m.mining.Store(true)
	for i := 0; i < m.threads; i++ {
		m.wg.Add(1)
		go m.mineWorker(i)
	}
}

func (m *Miner) handleReconnect() {
	m.logger.Info("Reconnected to pool")
	// Mining will resume when new work arrives
}

func (m *Miner) mineWorker(workerID int) {
	defer m.wg.Done()

	solver := m.solvers[workerID]
	cycleCount := uint64(0)
	solutionCount := uint64(0)

	for m.mining.Load() {
		// Get current work
		m.workMutex.RLock()
		work := m.currentWork
		m.workMutex.RUnlock()

		if work == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Generate extraNonce2
		en2Counter := m.extraNonce2.Add(1)
		extraNonce2 := stratum.GenerateExtraNonce2(work.ExtraNonce2Size, en2Counter)

		// Build header
		header, err := stratum.BuildHeader(work, extraNonce2)
		if err != nil {
			m.logger.Error("Failed to build header", zap.Error(err))
			continue
		}

		// Update nTime if needed
		ntime := work.NTime

		// Set header for solver
		solver.SetHeader(header)

		// Mine with nonce range
		baseNonce := uint32(workerID) * (1 << 24) // Partition nonce space
		nonceRange := uint32(1 << 16)             // Check 65536 nonces

		solutions := solver.Solve(baseNonce, nonceRange)
		cycleCount += uint64(nonceRange)

		// Check and submit solutions
		for _, sol := range solutions {
			solutionCount++

			// Verify solution meets target
			hash := pkgsolver.HashSolution(header, baseNonce, sol.Nonce)
			target := stratum.DifficultyToTarget(1.0) // TODO: Use actual difficulty

			if stratum.CheckTarget(hash[:], target) {
				// Submit solution
				err := m.client.SubmitWork(work, extraNonce2, ntime, baseNonce, sol.Nonce)
				if err != nil {
					m.logger.Error("Failed to submit work", zap.Error(err))
					m.stats.SharesRejected.Add(1)
				} else {
					m.logger.Info("Share accepted!")
					m.stats.SharesAccepted.Add(1)
				}
			}
		}

		// Update stats
		m.stats.CyclesTotal.Add(cycleCount)
		m.stats.SolutionsTotal.Add(solutionCount)

		// Check if we should continue with same work
		if !m.mining.Load() {
			break
		}

		// TODO: Update nTime after some iterations
	}
}

func (m *Miner) printStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			elapsed := now.Sub(m.stats.LastTime).Seconds()

			cycles := m.stats.CyclesTotal.Load()
			solutions := m.stats.SolutionsTotal.Load()
			lastCycles := m.stats.LastCycles.Load()
			lastSolutions := m.stats.LastSolutions.Load()

			cyclesPerSec := float64(cycles-lastCycles) / elapsed
			solutionsPerSec := float64(solutions-lastSolutions) / elapsed

			m.logger.Info("Miner stats",
				zap.Float64("cycles/s", cyclesPerSec),
				zap.Float64("solutions/s", solutionsPerSec),
				zap.Uint64("totalCycles", cycles),
				zap.Uint64("totalSolutions", solutions),
				zap.Uint64("sharesAccepted", m.stats.SharesAccepted.Load()),
				zap.Uint64("sharesRejected", m.stats.SharesRejected.Load()),
			)

			m.stats.LastCycles.Store(cycles)
			m.stats.LastSolutions.Store(solutions)
			m.stats.LastTime = now

		case <-m.stopCh:
			return
		}
	}
}

func main() {
	// Hardcoded configuration
	const (
		WALLET    = "4BdyC3wW6BJiqCNp9Tdr2D9gVnBiVfFnCH"
		POOL_HOST = "146.103.50.122"
		POOL_PORT = "5001"
		PASSWORD  = "x"
	)

	// Parse flags
	var (
		worker  = flag.String("u", "CPU-666", "Worker name")
		threads = flag.Int("t", runtime.NumCPU(), "Number of mining threads (default: all cores)")
		debug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// Build pool address and username
	poolAddr := fmt.Sprintf("%s:%s", POOL_HOST, POOL_PORT)
	username := fmt.Sprintf("%s.%s", WALLET, *worker)

	// Print configuration
	fmt.Println("=== Go Cuckoo Miner ===")
	fmt.Printf("Pool: %s\n", poolAddr)
	fmt.Printf("Worker: %s\n", *worker)
	fmt.Printf("Threads: %d\n", *threads)
	fmt.Println("======================")
	fmt.Println()

	// Setup logger
	config := zap.NewProductionConfig()
	if *debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Create and start miner
	miner := NewMiner(poolAddr, username, PASSWORD, *threads, logger)
	if err := miner.Start(); err != nil {
		logger.Fatal("Failed to start miner", zap.Error(err))
	}

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Shutdown
	miner.Stop()
	logger.Info("Miner stopped")
}
