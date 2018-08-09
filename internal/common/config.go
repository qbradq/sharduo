package common

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// Configuration represents a key-value configuration file
type Configuration map[string]string

// Config is the global configuration file
var Config Configuration

// NewConfigurationFromFile creates a new Configuration object with the values in the named file
func NewConfigurationFromFile(path string) Configuration {
	c := make(Configuration)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Fatalln(err)
	}

	r := bufio.NewReader(f)
	var lineNumber = 1
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		c.readLine(line, path, lineNumber)
		lineNumber++
	}

	return c
}

// GetString gets a configuration value as a string
func (c Configuration) GetString(key, defaultValue string) string {
	v, ok := c[key]
	if ok == false {
		v = defaultValue
	}
	return v
}

// GetInt gets a configuration value as an int safely
func (c Configuration) GetInt(key string, defaultValue int) int {
	v, ok := c[key]
	if ok == false {
		return defaultValue
	}
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(i)
}

func (c Configuration) readLine(line, path string, lineNumber int) {
	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		return
	}
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		log.Fatalf("%s:%d Bad config line, format is \"key=value\"", path, lineNumber)
	}
	_, duplicate := c[parts[0]]
	if duplicate {
		log.Fatalf("%s:%d Duplicate key \"%s\"", path, lineNumber, parts[0])
	}
	c[parts[0]] = parts[1]
}
