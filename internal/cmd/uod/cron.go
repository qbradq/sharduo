package uod

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/commands"
	"github.com/qbradq/sharduo/internal/configuration"
)

// cronJob represents one cron job.
type cronJob struct {
	// Minute of the hour to execute, -1 means all
	minute int
	// Hour of the day to execute, -1 means all
	hour int
	// Day of the week to execute, -1 means all
	day int
	// Command line to run
	command string
}

// cron is the global Cron object.
var cron Cron

// Cron reads the crontab file and executes commands at the prescribed time.
type Cron struct {
	// List of cron jobs
	jobs []cronJob
	// Done channel
	done chan struct{}
}

// InitializeCron loads the crontab into the global cron object.
func InitializeCron() error {
	// Initialize the cron structure
	cron.done = make(chan struct{})
	// Load the crontab or copy the default one
	d, err := os.ReadFile(configuration.CrontabFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d, err = writeDefaultCrontab()
			if err != nil {
				return err
			}
			// If we reach here we fall through to the rest of the function with
			// d set to the data of our crontab file
		} else {
			return err
		}
	}
	// Process crontab
	r := csv.NewReader(bytes.NewReader(d))
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = 4
	for {
		row, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if row == nil {
			break
		}
		m := -1
		if row[0] != "*" {
			n, err := strconv.ParseInt(row[0], 0, 32)
			if err != nil {
				return nil
			}
			m = int(n)
		}
		h := -1
		if row[1] != "*" {
			n, err := strconv.ParseInt(row[1], 0, 32)
			if err != nil {
				return nil
			}
			h = int(n)
		}
		d := -1
		if row[2] != "*" {
			n, err := strconv.ParseInt(row[2], 0, 32)
			if err != nil {
				return nil
			}
			d = int(n)
		}
		cron.jobs = append(cron.jobs, cronJob{
			minute:  m,
			hour:    h,
			day:     d,
			command: row[3],
		})
	}
	return nil
}

// writeDefaultCrontab writes out the default configuration file to
// configuration.CrontabFile.
func writeDefaultCrontab() ([]byte, error) {
	d, err := data.FS.ReadFile(configuration.DefaultCrontabFile)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(configuration.CrontabFile, d, 0777)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Main is the main loop for the Cron daemon.
func (c *Cron) Main(wg *sync.WaitGroup) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("panic: %v\n%s\n", p, debug.Stack())
			panic(p)
		}
	}()
	defer wg.Done()

	// Create the nil net state we use for the commands
	n := NewNetState(nil)

	// Ticker to check cron states every minute
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Ticks
	for {
		select {
		case t := <-ticker.C:
			t = t.Local()
			// Process all cron jobs
			for _, j := range c.jobs {
				// Make sure it's the correct time for the job to fire
				if j.minute >= 0 && j.minute != t.Minute() {
					continue
				}
				if j.hour >= 0 && j.hour != t.Hour() {
					continue
				}
				if j.day >= 0 && j.day != int(t.Weekday()) {
					continue
				}
				// Execute the command
				n.account = world.superUser
				commands.Execute(n, j.command)
			}
		case <-c.done:
			return
		}
	}
}

// Stop stops the cron daemon process.
func (c *Cron) Stop() {
	c.done <- struct{}{}
	close(c.done)
}
