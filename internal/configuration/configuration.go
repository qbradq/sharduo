package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

//
// Internal data paths
//

// Internal data file path for the default configuration ini
var DefaultConfigurationFile string = "misc/default-configuration.ini"

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
var ConfigurationFile string = "configuration.ini"

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

//
// Game configuration
//

// Starting location
var StartingLocation uo.Location

// Starting facing
var StartingFacing uo.Direction

// Load loads the configuration from the file
func Load() error {
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
			return nil
		}
	}
	// Read in the tag file object
	tfr := &util.TagFileReader{}
	tfr.StartReading(bytes.NewReader(d))
	tfo := tfr.ReadObject()
	if tfo == nil || tfr.HasErrors() {
		for _, err := range tfr.Errors() {
			log.Println(err)
		}
		return fmt.Errorf("%d errors while loading configuration", len(tfr.Errors()))
	}
	//
	// Read configuration values
	//
	// Internal paths
	TemplatesDirectory = tfo.GetString("TemplatesDirectory", "templates")
	ListsDirectory = tfo.GetString("ListsDirectory", "templates")
	TemplateVariablesFile = tfo.GetString("TemplateVariablesFile", "misc/templates")
	// External paths
	SaveDirectory = tfo.GetString("SaveDirectory", "saves")
	ArchiveDirectory = tfo.GetString("ArchiveDirectory", "archives")
	ClientFilesDirectory = tfo.GetString("ClientFilesDirectory", "client")
	CrontabFile = tfo.GetString("CrontabFile", "crontab")
	// Login service configuration
	LoginServerAddress = tfo.GetString("LoginServerAddress", "0.0.0.0")
	LoginServerPort = tfo.GetNumber("LoginServerPort", 7775)
	// Game service configuration
	GameServerAddress = tfo.GetString("GameServerAddress", "0.0.0.0")
	GameServerPublicAddress = tfo.GetString("GameServerPublicAddress", "127.0.0.1")
	GameServerPort = tfo.GetNumber("GameServerPort", 7777)
	GameSaveType = tfo.GetString("GameSaveType", "Flat")
	GameServerName = tfo.GetString("GameServerName", "ShardUO TC")
	// Debug flags
	GenerateDebugMaps = tfo.GetBool("GenerateDebugMaps", false)
	// Game configuration
	StartingLocation = tfo.GetLocation("StartingLocation", uo.Location{
		X: 0,
		Y: 0,
		Z: 0,
	})
	StartingFacing = uo.Direction(tfo.GetNumber("StartingFacing", 4))

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
