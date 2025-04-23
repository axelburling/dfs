package pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/axelburling/dfs/pkg/pubsub/grpc/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	// targetAddress     = "localhost:4981"
	poolSize          = 5
	keepAliveTimeout  = 5 * time.Minute
	connectionTimeout = 10 * time.Second
	idleTimeout       = 5 * time.Minute
	checkInterval     = 1 * time.Minute
	maxStreamsPerConn = 100
)

type Conn struct {
	*grpc.ClientConn
	pb.PublisherClient
	pb.SubscriberClient
	Name          string
	activeStreams int
	lastActivity  time.Time
	mu            sync.Mutex
}

type GrpcPool struct {
	conns []*Conn
	mu    sync.Mutex
	index int
	fn    CreateConnFn
}

type CreateConnFn func(string) (*Conn, error)

func New(target string, size int, fn CreateConnFn) (*GrpcPool, error) {
	pool := &GrpcPool{fn: fn}

	for i := 0; i < size; i++ {
		conn, err := fn(target)

		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}

		pool.conns = append(pool.conns, conn)
	}

	go pool.maintainPool(target)

	return pool, nil
}

func CreateGRPCConn(target string) (*Conn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                keepAliveTimeout,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.NewClient(target, opts...)

	if err != nil {
		return nil, err
	}

	c := Conn{
		ClientConn:       conn,
		Name:             uuid.New().String(),
		activeStreams:    0,
		lastActivity:     time.Now(),
		PublisherClient:  pb.NewPublisherClient(conn),
		SubscriberClient: pb.NewSubscriberClient(conn),
	}

	return &c, nil
}

func (p *GrpcPool) GetStream() (*Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) == 0 {
		return nil, fmt.Errorf("connection pool is empty")
	}

	startIndex := 0
	if p.index >= len(p.conns) {
		p.index = 0
	}

	startIndex = p.index

	for {
		conn := p.conns[p.index]

		conn.mu.Lock()

		if conn.activeStreams < maxStreamsPerConn {
			conn.activeStreams++
			conn.lastActivity = time.Now()
			conn.mu.Unlock()

			p.index = (p.index + 1) % len(p.conns)
			return conn, nil
		}
		conn.mu.Unlock()

		p.index = (p.index + 1) % len(p.conns)

		if p.index == startIndex {
			return nil, fmt.Errorf("no available streams across all connections")
		}
	}
}

func (p *GrpcPool) ReleaseConnection(conn *Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn.activeStreams--
	conn.lastActivity = time.Now()
}

func (p *GrpcPool) maintainPool(target string) {
	for {
		time.Sleep(checkInterval)

		p.mu.Lock()

		for i := len(p.conns) - 1; i >= 0; i-- {
			conn := p.conns[i]

			// Check for idle connections
			conn.mu.Lock()

			if conn.activeStreams == 0 && time.Since(conn.lastActivity) > idleTimeout {
				conn.Close()
				p.conns = append(p.conns[:i], p.conns[i+1:]...)
			}
			conn.mu.Unlock()
		}

		// Ensure pool size
		for len(p.conns) < poolSize {
			conn, err := p.fn(target)
			if err != nil {
				fmt.Printf("Failed to create new connection: %v\n", err)
				continue
			}
			conn.lastActivity = time.Now()
			p.conns = append(p.conns, conn)
		}

		p.mu.Unlock()
	}
}

func (p *GrpcPool) Close() {
	for _, conn := range p.conns {
		conn.Close()
	}
}
