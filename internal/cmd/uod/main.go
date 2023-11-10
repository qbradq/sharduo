package uod

import (
	"errors"
	"image"
	"image/png"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/qbradq/sharduo/internal/ai"
	"github.com/qbradq/sharduo/internal/commands"
	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/internal/events"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
)

// Tile data
var tiledatamul *file.TileDataMul

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
			log.Println("attempting last-ditch save from signal handler")
			wg, err := world.Marshal()
			if err != nil {
				log.Printf("last-ditch save from signal handler failed: %s", err.Error())
			} else {
				log.Printf("writing last-ditch save to disk")
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
	log.Println("ShardUO initializing...")

	// Load configuration
	if err := configuration.Load(); err != nil {
		log.Fatal(err)
	}

	// Load crontab
	if err := InitializeCron(); err != nil {
		log.Fatal(err)
	}

	// Load client data files
	log.Println("loading client files...")
	tiledatamul = file.NewTileDataMul(path.Join(configuration.ClientFilesDirectory, "tiledata.mul"))
	mapmul := file.NewMapMulFromFile(path.Join(configuration.ClientFilesDirectory, "map0.mul"), tiledatamul)
	staticsmul := file.NewStaticsMulFromFile(
		path.Join(configuration.ClientFilesDirectory, "staidx0.mul"),
		path.Join(configuration.ClientFilesDirectory, "statics0.mul"),
		tiledatamul)
	if tiledatamul == nil || mapmul == nil || staticsmul == nil {
		if tiledatamul == nil {
			log.Fatal("failed to load tiledata.mul")
		}
		if mapmul == nil {
			log.Fatal("failed to load map0.mul")
		}
		if staticsmul == nil {
			log.Fatal("failed to load statics0.mul")
		}
	}

	log.Println("Misc operations...")
	if configuration.GenerateDebugMaps {
		log.Println("generating debug map...")
		rcolmul := file.NewRadarColMulFromFile(path.Join(configuration.ClientFilesDirectory, "radarcol.mul"))
		if rcolmul == nil {
			log.Fatal("failed to load radarcol.mul")
		}
		rcols := rcolmul.Colors()
		mapimg := image.NewRGBA(image.Rect(0, 0, uo.MapWidth, uo.MapHeight))
		// Lay down the tiles
		for iy := 0; iy < uo.MapHeight; iy++ {
			for ix := 0; ix < uo.MapWidth; ix++ {
				t := mapmul.GetTile(ix, iy)
				mapimg.Set(ix, iy, rcols[t.BaseGraphic()])
			}
		}
		// Add statics
		for _, static := range staticsmul.Statics() {
			mapimg.Set(int(static.Location.X), int(static.Location.Y), rcols[static.BaseGraphic()+0x4000])
		}
		// Write out the map
		mapimgf, err := os.Create("debug-map.png")
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(mapimgf, mapimg); err != nil {
			log.Fatal(err)
		}
		mapimgf.Close()
	}

	// RNG initialization
	rng := util.NewRNG()

	// AI system initialization
	game.SetAIGetter(func(s string) game.AIModel {
		return ai.GetModel(s)
	})

	// Event system initialization
	game.SetEventHandlerGetter(func(which string) *game.EventHandler {
		return (*game.EventHandler)(events.GetEventHandler(which))
	})
	game.SetEventIndexGetter(func(which string) uint16 {
		return events.GetEventHandlerIndex(which)
	})

	// Command system initialization
	commands.RegisterCallbacks(
		GlobalChat,
		func() { world.Marshal() },
		Broadcast,
		gracefulShutdown,
		func() string { return world.LatestSavePath() })

	// Marshal system initialization
	marshal.SetInsertFunction(func(i interface{}) {
		o, ok := i.(game.Object)
		if !ok {
			return
		}
		world.Insert(o)
	})

	// Load object templates
	log.Println("Loading templates...")
	errs := template.Initialize(configuration.TemplatesDirectory,
		configuration.ListsDirectory, configuration.TemplateVariablesFile,
		rng, func(o template.Object) {
			if o == nil {
				return
			}
			obj, ok := o.(game.Object)
			if !ok {
				return
			}
			world.addNewObjectToDataStores(obj)
			obj.SetParent(game.TheVoid)
		})
	for _, err := range errs {
		log.Println(err)
	}
	if len(errs) > 0 {
		log.Fatalf("%d errors while loading object templates", len(errs))
	}

	// Initialize our data structures
	log.Println("Allocating world data structures...")
	world = NewWorld(configuration.SaveDirectory, rng)
	log.Println("Populating map data structures...")
	world.Map().LoadFromMuls(mapmul, staticsmul)
	game.RegisterWorld(world)

	// Inject server-side dynamic objects
	log.Println("Creating dynamic map objects...")

	// Try to load the most recent save
	if err := world.Unmarshal(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("warning: no save files found, executing first-start routine")
			firstStart()
		} else {
			log.Fatal("error while trying to load data stores from main goroutine", err)
		}
	}
}

// Executed on the first start of a new server.
func firstStart() {
	world.AuthenticateAccount("root", game.HashPassword("password"))
}

// Commands executed at every start.
func startCommands() {
	n := NewNetState(nil)
	n.account = world.superUser
	commands.Execute(n, "loadstatics")
	commands.Execute(n, "loaddoors")
	commands.Execute(n, "loadsigns")
	// This is really hacky, but the mobiles need to update what they are
	// standing on in this step and what they are standing on won't be there if
	// it's a dynamic static object created by [loadstatics and friends.
	world.m.AfterUnmarshal()
	// Create the spawners last as that forces a full respawn.
	commands.Execute(n, "loadspawners")
}

// Main is the entry point for uod.
func Main() {
	trap()
	initialize()
	startCommands()

	wg := &sync.WaitGroup{}

	// Start the goroutines
	wg.Add(4)
	go world.Main(wg)
	go cron.Main(wg)
	go LoginServerMain(wg)
	go GameServerMain(wg)
	wg.Wait()

	// Always save right before we go down
	wg, err := world.Marshal()
	if err != nil {
		log.Printf("ERROR SAVING WORLD AT END OF MAIN: %s", err.Error())
	} else {
		wg.Wait()
	}
}
