package uod

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/util"
)

// Configuration holds all of the configuration variables for the server.
type Configuration struct {

	//
	// Internal data paths
	//

	// Internal data file path for the default configuration ini
	DefaultConfigurationFile string
	// Internal data directory where templates are loaded from
	TemplatesDirectory string
	// Internal data directory where lists are loaded from
	ListsDirectory string
	// Internal data file path for the template variables
	TemplateVariablesFile string

	//
	// External data paths
	//

	// External path to the configuration file
	ConfigurationFile string
	// External directory path to write saves to
	SaveDirectory string
	// External directory containing the client files
	ClientFilesDirectory string

	//
	// Login service configuration
	//

	// IPv4 address to bind to
	LoginServerAddress string
	// TCP port to bind to
	LoginServerPort int

	//
	// Game service configuration
	//

	// IPv4 address to bind to
	GameServerAddress string
	// TCP port to bind to
	GameServerPort int
	// Save file type
	GameSaveType string

	//
	// Debug flags
	//

	// If true we should generate all of the debug maps at server start
	GenerateDebugMaps bool
}

// newConfiguration returns a new Configuration object
func newConfiguration() *Configuration {
	return &Configuration{
		DefaultConfigurationFile: "misc/default-configuration.ini",
		ConfigurationFile:        "configuration.ini",
	}
}

// LoadConfiguration loads the configuration from the file indicated in
// c.ConfigurationFile
func (c *Configuration) LoadConfiguration() error {
	d, err := os.ReadFile(c.ConfigurationFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d, err = c.writeDefaultConfiguration()
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
	c.TemplatesDirectory = tfo.GetString("TemplatesDirectory", "templates")
	c.ListsDirectory = tfo.GetString("ListsDirectory", "templates")
	c.TemplateVariablesFile = tfo.GetString("TemplateVariablesFile", "misc/templates")
	// External paths
	c.SaveDirectory = tfo.GetString("SaveDirectory", "saves")
	c.ClientFilesDirectory = tfo.GetString("ClientFilesDirectory", "client")
	// Login service configuration
	c.LoginServerAddress = tfo.GetString("LoginServerAddress", "0.0.0.0")
	c.LoginServerPort = tfo.GetNumber("LoginServerPort", 7775)
	// Game service configuration
	c.GameServerAddress = tfo.GetString("GameServerAddress", "0.0.0.0")
	c.GameServerPort = tfo.GetNumber("GameServerPort", 7777)
	c.GameSaveType = tfo.GetString("GameSaveType", "Flat")
	// Debug flags
	c.GenerateDebugMaps = tfo.GetBool("GenerateDebugMaps", false)

	return nil
}

// writeDefaultConfiguration writes out the default configuration file to
// c.ConfigurationFile
func (c *Configuration) writeDefaultConfiguration() ([]byte, error) {
	d, err := data.FS.ReadFile(c.DefaultConfigurationFile)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(c.ConfigurationFile, d, 0777)
	if err != nil {
		return nil, err
	}
	return d, nil
}
