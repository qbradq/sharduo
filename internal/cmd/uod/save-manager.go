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
}

// NewSaveManager returns a new SaveManager object.
func NewSaveManager(savePath string) *SaveManager {
	return &SaveManager{
		savePath: savePath,
	}
}

// Load initializes all global data managers from the latest save files.
func (m *SaveManager) Load() error {
	if !m.lock.TryLock() {
		return nil
	}
	defer m.lock.Unlock()

	if err := os.MkdirAll(m.savePath, 0777); err != nil {
		log.Println("error creating save directory", err)
		return err
	}

	accountManager = game.NewAccountManager()
	objectManager = game.NewObjectManager()

	return nil
}

// reportErrors logs all errors in the slice, then returns a single error with
// a summary report.
func (m *SaveManager) reportErrors(errs []error) error {
	for _, err := range errs {
		log.Println(err)
	}
	return fmt.Errorf("%d errors reported", len(errs))
}

// Save generates and writes all of the save files unless another save is in
// progress.
func (m *SaveManager) Save() error {
	if !m.lock.TryLock() {
		return nil
	}
	defer m.lock.Unlock()

	if err := os.MkdirAll(m.savePath, 0777); err != nil {
		log.Println("error creating save directory", err)
		return err
	}

	z, err := os.Create(path.Join(m.savePath, m.getFileName()+".zip"))
	if err != nil {
		return err
	}
	defer z.Close()

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
	if errs := objectManager.Save(f); errs != nil {
		return m.reportErrors(errs)
	}

	return nil
}

// getFileName returns the file name portion of the save path without extension.
func (m *SaveManager) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}
