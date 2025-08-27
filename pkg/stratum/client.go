// Package stratum implements Stratum mining protocol client
package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/bits"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Client represents a Stratum client connection
type Client struct {
	// Connection
	addr   string
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	// Authentication
	username string
	password string

	// State
	connected       atomic.Bool
	subscribed      atomic.Bool
	authorized      atomic.Bool
	extraNonce1     string
	extraNonce2Size int

	// Current work
	currentWork *Work
	workMutex   sync.RWMutex

	// Message handling
	msgID        atomic.Int64
	pending      map[int64]chan *Response
	pendingMutex sync.RWMutex

	// Callbacks
	onNewWork   func(*Work)
	onReconnect func()

	// Logging
	logger *zap.Logger

	// Channels
	stopCh chan struct{}
	workCh chan *Work

	// Difficulty
	difficulty float64
}

// Work represents mining job from pool
type Work struct {
	JobID           string
	PrevHash        string
	Coinbase1       string
	Coinbase2       string
	MerkleBranch    []string
	Version         string
	NBits           string
	NTime           string
	CleanJobs       bool
	ExtraNonce1     string
	ExtraNonce2Size int
	Target          []byte
	difficulty      float64
}

// Request represents JSON-RPC request
type Request struct {
	ID     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// Response represents JSON-RPC response
type Response struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *Error          `json:"error"`
}

// Notification represents server push notification
type Notification struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// Error represents JSON-RPC error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewClient creates a new Stratum client
func NewClient(addr, username, password string, logger *zap.Logger) *Client {
	return &Client{
		addr:     addr,
		username: username,
		password: password,
		logger:   logger,
		pending:  make(map[int64]chan *Response),
		stopCh:   make(chan struct{}),
		workCh:   make(chan *Work, 1),
	}
}

// SetWorkHandler sets callback for new work
func (c *Client) SetWorkHandler(handler func(*Work)) {
	c.onNewWork = handler
}

// SetReconnectHandler sets callback for reconnection
func (c *Client) SetReconnectHandler(handler func()) {
	c.onReconnect = handler
}

// Connect establishes connection to pool
func (c *Client) Connect() error {
	c.logger.Info("Connecting to pool", zap.String("addr", c.addr))

	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.connected.Store(true)

	// Start reader goroutine
	go c.readLoop()

	// Subscribe and authorize
	if err := c.subscribe(); err != nil {
		c.Close()
		return err
	}

	if err := c.authorize(); err != nil {
		c.Close()
		return err
	}

	c.logger.Info("Connected and authorized")
	return nil
}

// Close closes the connection
func (c *Client) Close() {
	c.connected.Store(false)
	close(c.stopCh)
	if c.conn != nil {
		c.conn.Close()
	}
}

// subscribe sends mining.subscribe
func (c *Client) subscribe() error {
	req := &Request{
		ID:     c.nextID(),
		Method: "mining.subscribe",
		Params: []interface{}{"go-miner/1.0"},
	}

	resp, err := c.call(req)
	if err != nil {
		return err
	}

	var result []json.RawMessage
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return err
	}

	if len(result) >= 2 {
		json.Unmarshal(result[1], &c.extraNonce1)
		json.Unmarshal(result[2], &c.extraNonce2Size)
		c.subscribed.Store(true)
		c.logger.Info("Subscribed",
			zap.String("extraNonce1", c.extraNonce1),
			zap.Int("extraNonce2Size", c.extraNonce2Size))
	}

	return nil
}

// authorize sends mining.authorize
func (c *Client) authorize() error {
	req := &Request{
		ID:     c.nextID(),
		Method: "mining.authorize",
		Params: []interface{}{c.username, c.password},
	}

	resp, err := c.call(req)
	if err != nil {
		return err
	}

	var result bool
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("authorization failed")
	}

	c.authorized.Store(true)
	c.logger.Info("Authorized", zap.String("username", c.username))
	return nil
}

// SubmitWork submits a found solution
func (c *Client) SubmitWork(work *Work, nonce2 string, nTime string, nonce uint32, solution []uint32) error {
	// Convert solution to comma-separated decimal string
	solStr := ""
	for i, s := range solution {
		if i > 0 {
			solStr += ","
		}
		solStr += fmt.Sprintf("%d", s)
	}

	req := &Request{
		ID:     c.nextID(),
		Method: "mining.submit",
		Params: []interface{}{
			c.username,
			work.JobID,
			nonce2,
			nTime,
			fmt.Sprintf("%08x", bits.ReverseBytes32(nonce)),
			solStr,
		},
	}

	resp, err := c.call(req)
	if err != nil {
		return err
	}

	var result bool
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("submission rejected")
	}

	c.logger.Info("Share accepted", zap.String("jobID", work.JobID))
	return nil
}

// GetWork returns current work
func (c *Client) GetWork() *Work {
	c.workMutex.RLock()
	defer c.workMutex.RUnlock()
	return c.currentWork
}

// readLoop reads messages from server
func (c *Client) readLoop() {
	for c.connected.Load() {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.logger.Error("Read error", zap.Error(err))
			c.handleDisconnect()
			return
		}

		// Log raw line for debugging
		c.logger.Debug("Stratum RAW", zap.String("line", line))

		// Try to parse as response
		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err == nil && resp.ID != 0 {
			c.handleResponse(&resp)
			continue
		}

		// Try to parse as notification
		var notif Notification
		if err := json.Unmarshal([]byte(line), &notif); err == nil {
			// Log notification body
			c.logger.Debug("Stratum NOTIFY", zap.String("method", notif.Method), zap.Any("params", notif.Params))
			c.handleNotification(&notif)
			continue
		}

		c.logger.Warn("Unknown message", zap.String("line", line))
	}
}

// handleResponse handles RPC response
func (c *Client) handleResponse(resp *Response) {
	c.pendingMutex.RLock()
	ch, ok := c.pending[resp.ID]
	c.pendingMutex.RUnlock()

	if ok {
		ch <- resp
		c.pendingMutex.Lock()
		delete(c.pending, resp.ID)
		c.pendingMutex.Unlock()
	}
}

// handleNotification handles server notification
func (c *Client) handleNotification(notif *Notification) {
	switch notif.Method {
	case "mining.notify":
		c.handleMiningNotify(notif.Params)
	case "mining.set_difficulty":
		c.handleSetDifficulty(notif.Params)
	case "client.reconnect":
		c.handleReconnectRequest(notif.Params)
	}
}

// handleMiningNotify processes new work
func (c *Client) handleMiningNotify(params []interface{}) {
	if len(params) < 9 {
		c.logger.Error("Invalid mining.notify params")
		return
	}

	// Wait briefly for subscribe response to populate extranonce1/size
	if c.extraNonce1 == "" || c.extraNonce2Size == 0 {
		for i := 0; i < 20; i++ { // up to ~1s total
			if c.extraNonce1 != "" && c.extraNonce2Size > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	work := &Work{
		JobID:           params[0].(string),
		PrevHash:        params[1].(string),
		Coinbase1:       params[2].(string),
		Coinbase2:       params[3].(string),
		Version:         params[5].(string),
		NBits:           params[6].(string),
		NTime:           params[7].(string),
		CleanJobs:       params[8].(bool),
		ExtraNonce1:     c.extraNonce1,
		ExtraNonce2Size: c.extraNonce2Size,
	}

	// Parse merkle branch
	if branches, ok := params[4].([]interface{}); ok {
		for _, branch := range branches {
			if b, ok := branch.(string); ok {
				work.MerkleBranch = append(work.MerkleBranch, b)
			}
		}
	}

	c.workMutex.Lock()
	c.currentWork = work
	c.workMutex.Unlock()

	if c.onNewWork != nil {
		c.onNewWork(work)
	}

	c.logger.Info("New work", zap.String("jobID", work.JobID))
}

// handleSetDifficulty updates difficulty
func (c *Client) handleSetDifficulty(params []interface{}) {
	if len(params) > 0 {
		if diff, ok := params[0].(float64); ok {
			c.difficulty = diff
			c.logger.Info("Difficulty set", zap.Float64("difficulty", diff))
			// target calculation is handled by miner using this value
		}
	}
}

// handleReconnectRequest handles pool reconnect request
func (c *Client) handleReconnectRequest(params []interface{}) {
	c.logger.Info("Reconnect requested by pool")
	go c.reconnect()
}

// handleDisconnect handles connection loss
func (c *Client) handleDisconnect() {
	c.connected.Store(false)
	c.logger.Warn("Disconnected from pool")
	go c.reconnect()
}

// reconnect attempts to reconnect
func (c *Client) reconnect() {
	for !c.connected.Load() {
		c.logger.Info("Attempting reconnect...")
		if err := c.Connect(); err != nil {
			c.logger.Error("Reconnect failed", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		if c.onReconnect != nil {
			c.onReconnect()
		}
		break
	}
}

// GetDifficulty returns the last set pool difficulty
func (c *Client) GetDifficulty() float64 {
	return c.difficulty
}

// call makes RPC call
func (c *Client) call(req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *Response, 1)
	c.pendingMutex.Lock()
	c.pending[req.ID] = ch
	c.pendingMutex.Unlock()

	if _, err := c.writer.Write(append(data, '\n')); err != nil {
		return nil, err
	}

	if err := c.writer.Flush(); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", resp.Error.Message)
		}
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("RPC timeout")
	}
}

// nextID generates next message ID
func (c *Client) nextID() int64 {
	return c.msgID.Add(1)
}
