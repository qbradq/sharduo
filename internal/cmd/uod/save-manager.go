package uod

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
)

// SaveManager manages the saving and loading of save data.
type SaveManager struct {
	// Root path of save data
	savePath string
	// Save lock
	lock sync.Mutex
	// World we are responsible for saving
	w *World
	// Account manager associated with this world
	am *game.AccountManager
}

// NewSaveManager returns a new SaveManager object.
func NewSaveManager(w *World, am *game.AccountManager, savePath string) *SaveManager {
	return &SaveManager{
		savePath: savePath,
		w:        w,
		am:       am,
	}
}

// reportErrors logs all errors in the slice, then returns a single error with
// a summary report.
func (m *SaveManager) reportErrors(errs []error) error {
	for _, err := range errs {
		log.Println(err)
	}
	return fmt.Errorf("%d errors reported", len(errs))
}

// Load reads in an entire save file.
func (m *SaveManager) Load() error {
	if !m.lock.TryLock() {
		return nil
	}
	defer m.lock.Unlock()

	// Look for the newest save and load it.
	entries, err := os.ReadDir(m.savePath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	latestIdx := -1
	var latestModTime time.Time
	for i, e := range entries {
		info, err := e.Info()
		if err != nil {
			return err
		}
		if info.ModTime().After(latestModTime) {
			latestModTime = info.ModTime()
			latestIdx = i
		}
	}
	e := entries[latestIdx]
	filePath := path.Join(m.savePath, e.Name())
	zf, err := zip.OpenReader(filePath)
	if err != nil {
		return nil
	}
	defer zf.Close()
	log.Println("loading data stores from", filePath)

	r, err := zf.Open("accounts.ini")
	if err != nil {
		return err
	}
	if errs := m.am.Load(r); errs != nil {
		r.Close()
		return m.reportErrors(errs)
	}
	r.Close()

	r, err = zf.Open("objects.ini")
	if err != nil {
		return err
	}
	if errs := world.Load(r); errs != nil {
		r.Close()
		return m.reportErrors(errs)
	}
	r.Close()

	return nil
}

// Save generates and writes all of the save files unless another save is in
// progress.
func (m *SaveManager) Save() error {
	if !m.lock.TryLock() {
		return nil
	}
	defer m.lock.Unlock()

	// Initialization in progress
	if accountManager == nil || world == nil {
		return nil
	}

	if err := os.MkdirAll(m.savePath, 0777); err != nil {
		log.Println("error creating save directory", err)
		return err
	}

	filePath := path.Join(m.savePath, m.getFileName()+".zip")
	z, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer z.Close()
	log.Println("saving data stores to", filePath)

	w := zip.NewWriter(z)
	defer w.Close()

	f, err := w.Create("accounts.ini")
	if err != nil {
		return err
	}
	if errs := accountManager.Save(f); errs != nil {
		return m.reportErrors(errs)
	}

	f, err = w.Create("objects.ini")
	if err != nil {
		return err
	}
	if errs := world.Save(f); errs != nil {
		return m.reportErrors(errs)
	}

	return nil
}

// getFileName returns the file name portion of the save path without extension.
func (m *SaveManager) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}
