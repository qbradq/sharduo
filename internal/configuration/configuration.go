package configuration

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/uo"
)

//
// Internal data paths
//

// Internal data file path for the default configuration ini
var DefaultConfigurationFile string = "misc/default-configuration.json"

// Internal data file path for the default crontab file
var DefaultCrontabFile string = "misc/default-crontab"

// Internal data directory where templates are loaded from
var TemplatesDirectory string

// Internal data directory where lists are loaded from
var ListsDirectory string

// Internal data file path for the template variables
var TemplateVariablesFile string

//
// External data paths
//

// External path to the configuration file
var ConfigurationFile string = "configuration.json"

// External directory path to write saves to
var SaveDirectory string

// External directory path to write archived saves to
var ArchiveDirectory string

// External directory containing the client files
var ClientFilesDirectory string

// External path to the crontab file
var CrontabFile string

//
// Login service configuration
//

// IPv4 address to bind to
var LoginServerAddress string

// TCP port to bind to
var LoginServerPort int

//
// Game service configuration
//

// IPv4 address to bind to
var GameServerAddress string

// IPv4 address to give to clients to connect to the game server
var GameServerPublicAddress string

// TCP port to bind to
var GameServerPort int

// Save file type
var GameSaveType string

// Name of the game server
var GameServerName string

//
// Debug flags
//

// If true we should generate all of the debug maps at server start
var GenerateDebugMaps bool

// If true we should enter CPU profiling mode for the main server loop
var CPUProfile bool

//
// Game configuration
//

// Starting location
var StartingLocation uo.Point

// Starting facing
var StartingFacing uo.Direction

// Load loads the configuration from the file
func Load() error {
	dict := map[string]string{}
	GetString := func(n, d string) string {
		if v, found := dict[n]; found {
			return v
		}
		return d
	}
	GetNumber := func(n string, d int) int {
		if s, found := dict[n]; found {
			v, err := strconv.ParseInt(s, 0, 32)
			if err != nil {
				panic(err)
			}
			return int(v)
		}
		return d
	}
	GetBool := func(n string, d bool) bool {
		if s, found := dict[n]; found {
			v, err := strconv.ParseBool(s)
			if err != nil {
				panic(err)
			}
			return v
		}
		return d
	}
	GetPoint := func(n string, d uo.Point) uo.Point {
		ret := d
		if s, found := dict[n]; found {
			parts := strings.Split(s, ",")
			if len(parts) != 3 {
				panic("expected three parts")
			}
			v, err := strconv.ParseInt(parts[0], 0, 32)
			if err != nil {
				panic(err)
			}
			ret.X = int(v)
			v, err = strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				panic(err)
			}
			ret.Y = int(v)
			v, err = strconv.ParseInt(parts[2], 0, 32)
			if err != nil {
				panic(err)
			}
			ret.Z = int(v)
		}
		return ret
	}
	d, err := os.ReadFile(ConfigurationFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d, err = writeDefaultConfiguration()
			if err != nil {
				return err
			}
			// If we reach here we fall through to the rest of the function with
			// d set to the data of our configuration file
		} else {
			return err
		}
	}
	// Read in the json file containing all of the configuration options
	if err := json.Unmarshal(d, &dict); err != nil {
		return err
	}
	//
	// Read configuration values
	//
	// Internal paths
	TemplatesDirectory = GetString("TemplatesDirectory", "templates")
	ListsDirectory = GetString("ListsDirectory", "templates")
	TemplateVariablesFile = GetString("TemplateVariablesFile", "misc/templates")
	// External paths
	SaveDirectory = GetString("SaveDirectory", "saves")
	ArchiveDirectory = GetString("ArchiveDirectory", "archives")
	ClientFilesDirectory = GetString("ClientFilesDirectory", "client")
	CrontabFile = GetString("CrontabFile", "crontab")
	// Login service configuration
	LoginServerAddress = GetString("LoginServerAddress", "0.0.0.0")
	LoginServerPort = GetNumber("LoginServerPort", 7775)
	// Game service configuration
	GameServerAddress = GetString("GameServerAddress", "0.0.0.0")
	GameServerPublicAddress = GetString("GameServerPublicAddress", "127.0.0.1")
	GameServerPort = GetNumber("GameServerPort", 7777)
	GameSaveType = GetString("GameSaveType", "Flat")
	GameServerName = GetString("GameServerName", "ShardUO TC")
	// Debug flags
	GenerateDebugMaps = GetBool("GenerateDebugMaps", false)
	CPUProfile = GetBool("CPUProfile", false)
	// Game configuration
	StartingLocation = GetPoint("StartingLocation", uo.Point{
		X: 0,
		Y: 0,
		Z: 0,
	})
	StartingFacing = uo.Direction(GetNumber("StartingFacing", 4))

	return nil
}

// writeDefaultConfiguration writes out the default configuration file
func writeDefaultConfiguration() ([]byte, error) {
	d, err := data.FS.ReadFile(DefaultConfigurationFile)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(ConfigurationFile, d, 0777)
	if err != nil {
		return nil, err
	}
	return d, nil
}
