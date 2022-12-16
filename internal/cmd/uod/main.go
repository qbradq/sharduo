package uod

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Map of all active net states
var netStates sync.Map

// The world we are running
var world *World

// The template manager
var templateManager *TemplateManager

// The configuration
var configuration *Configuration

// trap is used to trap all of the system signals.
func trap(l *net.TCPListener) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			// Close the main listener and let the graceful shutdown save the
			// data stores.
			l.Close()
		} else {
			// Last-ditch save attempt
			log.Println("attempting last-ditch save from signal handler...")
			if err := world.Save(); err != nil {
				log.Fatalf("last-ditch save from signal handler failed:%v\n", err)
			}
		}
	}()
}

// Main is the entry point for uod.
func Main() {
	// Load configuration
	configuration = newConfiguration()
	if err := configuration.LoadConfiguration(); err != nil {
		log.Fatal(err)
	}

	// TODO Load client data files
	// Debug
	// log.Println("loading client files...")
	// rcolmul := file.NewRadarColMulFromFile(path.Join(configuration.ClientFilesDirectory, "radarcol.mul"))
	// mapmul := file.NewMapMulFromFile(path.Join(configuration.ClientFilesDirectory, "map0.mul"))
	// if rcolmul == nil || mapmul == nil {
	// 	os.Exit(1)
	// }
	// log.Println("generating debug map...")
	// rcols := rcolmul.Colors()
	// mapimg := image.NewRGBA(image.Rect(0, 0, uo.MapWidth, uo.MapHeight))
	// for iy := 0; iy < uo.MapHeight; iy++ {
	// 	for ix := 0; ix < uo.MapWidth; ix++ {
	// 		t := mapmul.GetTile(ix, iy)
	// 		mapimg.Set(ix, iy, rcols[t.Graphic])
	// 	}
	// }
	// mapimgf, err := os.Create("debug-map.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err := png.Encode(mapimgf, mapimg); err != nil {
	// 	log.Fatal(err)
	// }
	// mapimgf.Close()

	// RNG initialization
	rng := util.NewRNG()

	// Load object templates
	templateManager = NewTemplateManager("templates")
	errs := templateManager.LoadAll(configuration.TemplatesDirectory, configuration.ListsDirectory)
	for _, err := range errs {
		log.Println(err)
	}
	if len(errs) > 0 {
		log.Fatalf("%d errors while loading object templates", len(errs))
	}

	// Initialize our data structures
	world = NewWorld(configuration.SaveDirectory, rng)
	game.RegisterWorld(world)

	// Try to load the most recent save
	if err := world.Load(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("warning: no save files found")
		} else {
			log.Println("error while trying to load data stores from main goroutine", err)
			os.Exit(1)
		}
	}

	// Debug
	for i := 0; i < 50; i++ {
		o := world.New(templateManager.NewObject("FancyShirt"))
		l := o.Location()
		l.X += i
		o.SetLocation(l)
		world.Map().AddNewObject(o)
	}

	// Start the goroutines
	go world.Process()
	go LoginServerMain()

	ipstr := "127.0.0.1"
	port := 7777

	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ipstr),
		Port: port,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("game server listening at %s:%d\n", ipstr, port)
	trap(l)
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			if err := world.Save(); err != nil {
				log.Println("error while trying to save data stores from main goroutine", err)
				os.Exit(1)
			}
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatal(err)
				return
			}
		}
		go handleConnection(c)
	}
}

// Goroutine for handling inbound connections.
func handleConnection(c *net.TCPConn) {
	ns := NewNetState(c)
	netStates.Store(ns, true)
	ns.Service()
	netStates.Delete(ns)
}

// Broadcast sends a system-wide broadcast message to all connected clients.
func Broadcast(format string, args ...interface{}) {
	s := "System Broadcast: " + fmt.Sprintf(format, args...)
	netStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.SystemMessage(s)
		return true
	})
}

// GlobalChat sends a global chat message to all connected clients.
func GlobalChat(who, text string) {
	s := fmt.Sprintf("%s: %s", who, text)
	netStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.Send(&serverpacket.Speech{
			Speaker: uo.SerialSystem,
			Body:    uo.BodySystem,
			Font:    uo.FontNormal,
			Hue:     1166,
			Name:    "",
			Text:    s,
			Type:    uo.SpeechTypeSystem,
		})
		return true
	})
}
