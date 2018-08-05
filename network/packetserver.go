package network

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// PacketServer represents an Ultima Online protocol server
type PacketServer struct {
	host        string
	port        int
	clients     map[*packetClient]struct{}
	clientsLock sync.RWMutex
	listener    *net.TCPListener
	wg          *sync.WaitGroup
}

// NewPacketServer creates a new PacketServer implementation
func NewPacketServer(host string, port int) *PacketServer {
	return &PacketServer{
		host:    host,
		port:    port,
		clients: make(map[*packetClient]struct{}),
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
		log.Panic(err)
	}
	p.listener = l
	defer p.listener.Close()
	log.Println("Listening on", p.listener.Addr().String())

	// Main loop
	for {
		conn, err := p.listener.AcceptTCP()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else {
				log.Println("Listen socket closing because of", err.Error())
				break
			}
		}
		p.addClient(conn)
	}
}

// Stop signals the server goroutine and all client goroutines to exit
func (p *PacketServer) Stop() {
	p.listener.Close()
	for client := range p.clients {
		client.Stop()
	}
}

// Add a client to the server
func (p *PacketServer) addClient(conn *net.TCPConn) {
	log.Println("Client connected from", conn.RemoteAddr().String())

	// Add the client
	client := newPacketClient(conn)
	p.clientsLock.Lock()
	p.clients[client] = struct{}{}
	p.clientsLock.Unlock()

	// Start client goroutines
	go func() {
		client.ReadLoop(p.wg)

		// Remove the client
		client.Stop()
		p.clientsLock.Lock()
		delete(p.clients, client)
		p.clientsLock.Unlock()
	}()
	go func() {
		client.WriteLoop(p.wg)

		// Remove the client
		client.Stop()
		p.clientsLock.Lock()
		delete(p.clients, client)
		p.clientsLock.Unlock()
	}()
}
