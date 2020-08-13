package uod

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/pkg/uo"
)

const (
	netReadBufLen       int           = 64 * 1024
	netWriteBufLen      int           = 128 * 1024
	netCompressBufLen   int           = 128 * 1024
	netKeepAliveTimeout time.Duration = time.Second * 5
	netReadTimeout      time.Duration = time.Minute * 15
)

var netConns chan *net.TCPConn
var netListener *net.TCPListener
var netWaitGroup *sync.WaitGroup
var netActiveConns map[*net.TCPConn]*netState
var netActiveConnsLock sync.Mutex

func netStart() {
	netConns = make(chan *net.TCPConn, 100)
	netWaitGroup = &sync.WaitGroup{}
	netActiveConns = make(map[*net.TCPConn]*netState)
	netListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 2593,
	})
	if err != nil {
		log.Fatal(err)
	}

	go netNewConns()

	for {
		// netListener.SetDeadline(time.Now().Add(netReadTimeout))
		conn, err := netListener.AcceptTCP()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") ||
				err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else {
				log.Println("Listen socket closing because of", err.Error())
				break
			}
		}
		select {
		case netConns <- conn:
			// Ignore
		default:
			log.Fatal("Channel overflow on netConns")
		}
	}
}

func netStop() {
	netListener.Close()
	netActiveConnsLock.Lock()
	close(netConns)
	for conn, ns := range netActiveConns {
		netCloseActiveConn(conn, ns)
	}
	netActiveConns = nil
	netActiveConnsLock.Unlock()
	netWaitGroup.Wait()
}

func netNewConns() {
	netWaitGroup.Add(1)
	defer netWaitGroup.Done()

	for conn := range netConns {
		ns := newNetState()
		netActiveConnsLock.Lock()
		netActiveConns[conn] = ns
		netActiveConnsLock.Unlock()
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(netKeepAliveTimeout)
		conn.SetLinger(0)
		conn.SetNoDelay(true)
		conn.SetReadBuffer(netReadBufLen)
		conn.SetWriteBuffer(netWriteBufLen)
		go netReadConn(conn, ns)
		go netWriteConn(conn, ns)
	}
}

func netCloseActiveConn(conn *net.TCPConn, ns *netState) {
	if ns.outboundPackets != nil {
		close(ns.outboundPackets)
		ns.outboundPackets = nil
	}
	conn.Close()
}

func netRemoveActiveConn(conn *net.TCPConn, ns *netState) {
	netCloseActiveConn(conn, ns)
	netActiveConnsLock.Lock()
	delete(netActiveConns, conn)
	netActiveConnsLock.Unlock()
}

func netReadConn(conn *net.TCPConn, ns *netState) {
	netWaitGroup.Add(1)
	defer netWaitGroup.Done()

	r := uo.NewClientPacketReader(conn)
	conn.SetDeadline(time.Now().Add(netReadTimeout))
	r.ReadConnectionHeader()
	for {
		conn.SetDeadline(time.Now().Add(netReadTimeout))
		p, err := r.ReadClientPacket()
		fatal, logged := netLogClientError(conn, p, err)
		if fatal || p == nil {
			break
		}
		cmd := p.Command()
		handler := clientPacketHandlers[cmd]
		if handler != nil {
			handler(ns, p)
		} else if logged == false {
			log.Printf("Client %s sent unhandled packet 0x%02X\n",
				conn.RemoteAddr().String(),
				p.Command())
		}
	}
	log.Printf("Client %s read disconnect\n", conn.RemoteAddr().String())
	netRemoveActiveConn(conn, ns)
}

func netWriteConn(conn *net.TCPConn, ns *netState) {
	netWaitGroup.Add(1)
	defer netWaitGroup.Done()

	cbuf := make([]byte, 0, netCompressBufLen)
	var obuf []byte

	for p := range ns.outboundPackets {
		if ns.CompressOutput() {
			// log.Printf("(compressed) %#v\n", p[:len(p)])
			cbuf = cbuf[:0]
			obuf = uo.HuffmanEncodePacket(p, cbuf)
		} else {
			// log.Printf("%#v\n", p[:len(p)])
			obuf = p
		}
		_, err := conn.Write(obuf)
		if fatal, _ := netLogClientError(conn, nil, err); fatal {
			break
		}
	}
	log.Printf("Client %s write disconnect\n", conn.RemoteAddr().String())
	netRemoveActiveConn(conn, ns)
}

func netLogClientError(c *net.TCPConn, packet interface{}, err error) (bool, bool) {
	if packet != nil {
		switch pkt := packet.(type) {
		case uo.ClientPacketInvalid:
			log.Printf("Client %s disconnected due to invalid packet 0x%02X\n",
				c.RemoteAddr().String(),
				pkt.Command())
			return true, true
		case uo.ClientPacketNotSupported:
			log.Printf("Client %s sent unsupported packet 0x%02X\n",
				c.RemoteAddr().String(),
				pkt.Command())
			return false, true
		}
	}
	if err == nil {
		return false, false
	} else if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return false, false
	} else if strings.Contains(err.Error(), "use of closed network connection") ||
		err == io.EOF || err == io.ErrUnexpectedEOF {
		return true, false
	}
	log.Println("Client",
		c.RemoteAddr().String(),
		"disconnected due to",
		err)
	return true, true
}
