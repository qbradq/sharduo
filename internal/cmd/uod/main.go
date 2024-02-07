package uod

import (
	"errors"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/pkg/profile"
	"github.com/qbradq/sharduo/internal/commands"
	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Tile data
var tileDataMul *file.TileDataMul

// The world we are running
var world *World

// gracefulShutdown initiates a graceful systems shutdown
func gracefulShutdown() {
	StopLoginService()
	StopGameService()
	cron.Stop()
	world.Stop()
}

// trap is used to trap all of the system signals.
func trap() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			gracefulShutdown()
		} else {
			// Last-ditch save attempt
			log.Println("warning: attempting last-ditch save from signal handler")
			wg, err := world.Marshal()
			if err != nil {
				log.Printf("error: last-ditch save from signal handler failed: %s", err.Error())
			} else {
				log.Printf("info: writing last-ditch save to disk")
				wg.Wait()
			}
			os.Exit(0)
		}
	}()
}

// initialize takes care of all of the memory-intensive initialization stuff
// so the main routine can let go of all the memory.
func initialize() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   "./logs/sharduo.log",
		MaxSize:    128,
		MaxAge:     28,
		MaxBackups: 3,
	}))
	log.Println("info: ShardUO initializing...")
	// Load configuration
	if err := configuration.Load(); err != nil {
		log.Fatal(err)
	}
	// Load crontab
	if err := InitializeCron(); err != nil {
		log.Fatal(err)
	}
	// Load client data files
	log.Println("info: loading client files")
	tileDataMul = file.NewTileDataMul(path.Join(configuration.ClientFilesDirectory, "tiledata.mul"))
	mapMul := file.NewMapMulFromFile(path.Join(configuration.ClientFilesDirectory, "map0.mul"), tileDataMul)
	staticsMul := file.NewStaticsMulFromFile(
		path.Join(configuration.ClientFilesDirectory, "staidx0.mul"),
		path.Join(configuration.ClientFilesDirectory, "statics0.mul"),
		tileDataMul)
	if tileDataMul == nil || mapMul == nil || staticsMul == nil {
		if tileDataMul == nil {
			log.Fatal("failed to load tiledata.mul")
		}
		if mapMul == nil {
			log.Fatal("failed to load map0.mul")
		}
		if staticsMul == nil {
			log.Fatal("failed to load statics0.mul")
		}
	}
	// Generate debug stuffs if requested
	log.Println("info: misc startup operations")
	if configuration.GenerateDebugMaps {
		log.Println("debug: generating debug map...")
		rColMul := file.NewRadarColMulFromFile(path.Join(configuration.ClientFilesDirectory, "radarcol.mul"))
		if rColMul == nil {
			log.Fatal("failed to load radarcol.mul")
		}
		rCols := rColMul.Colors()
		mapImg := image.NewRGBA(image.Rect(0, 0, uo.MapWidth, uo.MapHeight))
		// Lay down the tiles
		for iy := 0; iy < uo.MapHeight; iy++ {
			for ix := 0; ix < uo.MapWidth; ix++ {
				t := mapMul.GetTile(ix, iy)
				mapImg.Set(ix, iy, rCols[t.BaseGraphic()])
			}
		}
		// Add statics
		for _, static := range staticsMul.Statics() {
			mapImg.Set(int(static.Location.X), int(static.Location.Y), rCols[static.BaseGraphic()+0x4000])
		}
		// Write out the map
		mapImgF, err := os.Create("debug-map.png")
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(mapImgF, mapImg); err != nil {
			log.Fatal(err)
		}
		mapImgF.Close()
	}
	// Command system initialization
	commands.RegisterCallbacks(
		GlobalChat,
		func() { world.Marshal() },
		Broadcast,
		gracefulShutdown,
		func() string { return world.LatestSavePath() })
	// Initialize our data structures
	log.Println("info: allocating world data structures")
	world = NewWorld(configuration.SaveDirectory)
	game.World = world
	log.Println("info: populating map data structures")
	world.Map().LoadFromMuls(mapMul, staticsMul)
	// Load prototypes
	game.LoadItemPrototypes()
	game.LoadMobilePrototypes()
	// Inject server-side dynamic objects
	log.Println("info: creating dynamic map objects")
	// Try to load the most recent save
	if err := world.Unmarshal(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("warning: no save files found, executing first-start routine")
			firstStart()
		} else {
			log.Fatal("error: while trying to load data stores from main goroutine", err)
		}
	}
}

// Executed on the first start of a new server.
func firstStart() {
	world.AuthenticateAccount("root", util.HashPassword("password"))
}

// Commands executed at every start.
func startCommands() {
	n := NewNetState(nil)
	n.account = world.superUser
	commands.Execute(n, "load_regions")
	commands.Execute(n, "load_statics")
	commands.Execute(n, "load_doors")
	commands.Execute(n, "load_signs")
	commands.Execute(n, "respawn")
	// This is really hacky, but the mobiles need to update what they are
	// standing on in this step and what they are standing on won't be there if
	// it's a dynamic static object created by [load_statics and friends.
	world.m.AfterUnmarshal()
}

// Main is the entry point for uod.
func Main() {
	if configuration.LogPanics {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("panic: %v\n%s\n", p, debug.Stack())
				panic(p)
			}
		}()
	}
	trap()
	initialize()
	startCommands()

	wg := &sync.WaitGroup{}

	// Start the goroutines
	var ps interface{ Stop() }
	if configuration.CPUProfile {
		ps = profile.Start(profile.ProfilePath("."))
	}
	wg.Add(4)
	go world.Main(wg)
	go cron.Main(wg)
	go LoginServerMain(wg)
	go GameServerMain(wg)
	wg.Wait()
	if configuration.CPUProfile {
		ps.Stop()
	}

	// Always save right before we go down
	wg, err := world.Marshal()
	if err != nil {
		log.Printf("error: saving world at end of main: %s", err.Error())
	} else {
		wg.Wait()
	}
}
