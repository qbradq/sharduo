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
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/game/events"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
)

// Tile data
var tiledatamul *file.TileDataMul

// The world we are running
var world *World

// The template manager
var templateManager *TemplateManager

// The configuration
var configuration *Configuration

// trap is used to trap all of the system signals.
func trap() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			// Try graceful shutdown
			StopLoginService()
			StopGameService()
			world.Stop()
		} else {
			// Last-ditch save attempt
			log.Println("attempting last-ditch save from signal handler...")
			if err := world.Save(); err != nil {
				log.Printf("last-ditch save from signal handler failed: %s", err.Error())
			}
			os.Exit(0)
		}
	}()
}

// Initialize takes care of all of the memory-intensive initialization stuff
// so the main routine can let go of all the memory.
func Initialize() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Load configuration
	configuration = newConfiguration()
	if err := configuration.LoadConfiguration(); err != nil {
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
			mapimg.Set(static.Location.X, static.Location.Y, rcols[static.BaseGraphic()+0x4000])
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

	// Event system initialization
	game.SetEventHandlerGetter(func(which string) *func(game.Object, game.Object) {
		return events.GetEventHandler(which)
	})

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
	world.Map().LoadFromMuls(mapmul, staticsmul)
	game.RegisterWorld(world)

	// Try to load the most recent save
	if err := world.Unmarshal(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("warning: no save files found")
		} else {
			log.Fatal("error while trying to load data stores from main goroutine", err)
		}
	}
}

// Main is the entry point for uod.
func Main() {
	trap()
	Initialize()

	wg := &sync.WaitGroup{}

	// Start the goroutines
	go world.Main(wg)
	go LoginServerMain(wg)
	go GameServerMain(wg)
	time.Sleep(time.Second * 1)
	wg.Wait()
	if err := world.Marshal(); err != nil {
		log.Printf("ERROR SAVING WORLD AT END OF MAIN: %s", err.Error())
	}
}
