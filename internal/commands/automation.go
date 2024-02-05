package commands

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/internal/game"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Commands used for automation and crontab are placed here

func init() {
	reg(&cmDesc{"logMemStats", nil, commandLogMemStats, game.RoleAdministrator, "logMemStats", "Forces the server to log memory statistics and echo that to the caller"})
	reg(&cmDesc{"snapshot_clean", nil, commandSnapshotClean, game.RoleAdministrator, "snapshot_clean", "internal command, please do not use"})
	reg(&cmDesc{"snapshot_daily", nil, commandSnapshotDaily, game.RoleAdministrator, "snapshot_daily", "internal command, please do not use"})
	reg(&cmDesc{"snapshot_weekly", nil, commandSnapshotWeekly, game.RoleAdministrator, "snapshot_weekly", "internal command, please do not use"})
}

func commandLogMemStats(n game.NetState, args CommandArgs, cl string) {
	mb := func(n uint64) uint64 {
		return n / 1024 / 1024
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	p := message.NewPrinter(language.English)
	s := p.Sprintf("stats: HeapAlloc=%dMB, LiveObjects=%d", mb(m.HeapAlloc),
		m.Mallocs-m.Frees)
	log.Println(s)
	if n != nil && n.Mobile() != nil {
		n.Speech(nil, s)
	}
}

func commandSnapshotClean(n game.NetState, args CommandArgs, cl string) {
	t := time.Now().Add(time.Hour * 72 * -1)
	filepath.WalkDir(configuration.SaveDirectory, func(p string, d fs.DirEntry, e error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		di, err := d.Info()
		if err != nil {
			return err
		}
		if di.ModTime().Before(t) {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
		return nil
	})
	n.Speech(nil, "old saves cleaned")
}

func commandSnapshotDaily(n game.NetState, args CommandArgs, cl string) {
	// Make sure the archive directory exists
	os.MkdirAll(configuration.ArchiveDirectory, 0777)
	// Create an archive copy of the save file
	p := latestSavePath()
	src, err := os.Open(p)
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	defer src.Close()
	dest, err := os.Create(path.Join(configuration.ArchiveDirectory, "daily.sav.gz"))
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	_, err = io.Copy(dest, src)
	dest.Close()
	if err != nil {
		n.Speech(nil, "error: failed to create daily archive: %s", err)
		return
	}
	// Remove the oldest save
	os.Remove(path.Join(configuration.ArchiveDirectory, "daily7.sav.gz"))
	// Rotate daily saves
	for i := 6; i > 0; i-- {
		op := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("daily%d.sav.gz", i))
		np := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("daily%d.sav.gz", i+1))
		os.Rename(op, np)
	}
	// Move the new save file into the archives
	os.Rename(path.Join(configuration.ArchiveDirectory, "daily.sav.gz"),
		path.Join(configuration.ArchiveDirectory, "daily1.sav.gz"))
	n.Speech(nil, "daily archive complete")
}

func commandSnapshotWeekly(n game.NetState, args CommandArgs, cl string) {
	// Make sure the archive directory exists
	os.MkdirAll(configuration.ArchiveDirectory, 0777)
	// Create an archive copy of the save file
	p := latestSavePath()
	src, err := os.Open(p)
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	defer src.Close()
	dest, err := os.Create(path.Join(configuration.ArchiveDirectory, "weekly.sav.gz"))
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	_, err = io.Copy(dest, src)
	dest.Close()
	if err != nil {
		n.Speech(nil, "error: failed to create weekly archive: %s", err)
		return
	}
	// Remove the oldest save
	os.Remove(path.Join(configuration.ArchiveDirectory, "weekly52.sav.gz"))
	// Rotate daily saves
	for i := 51; i > 0; i-- {
		op := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("weekly%d.sav.gz", i))
		np := path.Join(configuration.ArchiveDirectory, fmt.Sprintf("weekly%d.sav.gz", i+1))
		os.Rename(op, np)
	}
	// Move the new save file into the archives
	os.Rename(path.Join(configuration.ArchiveDirectory, "weekly.sav.gz"),
		path.Join(configuration.ArchiveDirectory, "weekly1.sav.gz"))
	n.Speech(nil, "weekly archive complete")
}
