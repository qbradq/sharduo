package network

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/qbradq/sharduo/core"
	"github.com/satori/go.uuid"
)

const acceptTimeout = time.Second * 5
const keepaliveTimeout = time.Second * 5
const readBufferSize = 1024 * 128  // 128kb OS read buffer per client
const writeBufferSize = 1024 * 384 // 384kb OS write buffer per client

// PacketServer represents an Ultima Online protocol server
type PacketServer struct {
	host        string
	port        int
	clients     map[uuid.UUID]*PacketClient
	clientsLock sync.RWMutex
	listener    *net.TCPListener
	wg          *sync.WaitGroup
	core.Stopper
}

// NewPacketServer creates a new PacketServer implementation
func NewPacketServer(host string, port int) *PacketServer {
	return &PacketServer{
		host:    host,
		port:    port,
		clients: make(map[uuid.UUID]*PacketClient),
	}
}

// Run executes the main loop, start as a goroutine
func (p *PacketServer) Run(wg *sync.WaitGroup) {
	// WaitGroup stuff
	p.wg = wg
	p.wg.Add(1)
	defer p.wg.Done()

	// Create the listener
	addr := net.TCPAddr{
		IP:   net.ParseIP(p.host),
		Port: p.port,
	}
	l, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Panicln(err.Error())
		os.Exit(1)
	}
	p.listener = l
	defer p.listener.Close()
	log.Println("Listening on", p.listener.Addr().String())

	// Main loop
	for p.Stopping() == false {
		p.listener.SetDeadline(time.Now().Add(acceptTimeout))
		conn, err := p.listener.AcceptTCP()
		if err != nil {
			if noe, ok := err.(*net.OpError); ok && noe.Timeout() {
				continue
			} else {
				log.Fatalln("Listen socket closing because of", err.Error())
				break
			}
		}
		if p.Stopping() {
			continue
		}
		p.addClient(conn)
	}
}

// Stop signals the server goroutine and all client goroutines to exit
func (p *PacketServer) Stop() {
	p.Stopper.Stop()
	for _, k := range p.clients {
		k.Stop()
	}
}

// Add a client to the server
func (p *PacketServer) addClient(conn *net.TCPConn) {
	log.Println("Client connected from", conn.RemoteAddr().String())

	// Configure connection
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(keepaliveTimeout)
	conn.SetLinger(0)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(readBufferSize)
	conn.SetWriteBuffer(writeBufferSize)

	// Add the client
	client := &PacketClient{
		Conn: conn,
	}
	p.clientsLock.Lock()
	p.clients[client.UUID()] = client
	p.clientsLock.Unlock()
	go func() {
		p.wg.Add(1)
		client.ReadLoop()
		p.clientsLock.Lock()
		delete(p.clients, client.UUID())
		p.clientsLock.Unlock()
		client.Conn.Close()
		p.wg.Done()
	}()
}
