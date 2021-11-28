package runner

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && isWatchedFile(event.Name) {
					watcherLog("sending event %s", event)
					startChannel <- event.String()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				watcherLog("error: %s", err)
			}
		}
	}()

	wfolders := []string{}
	for _, p := range watchPaths() {
		tp := strings.TrimSpace(p)
		filepath.Walk(tp, func(pth string, info os.FileInfo, err error) error {
			if info.IsDir() && !isTmpDir(pth) {
				if len(pth) > 1 && strings.HasPrefix(filepath.Base(pth), ".") {
					return filepath.SkipDir
				}

				if isIgnoredFolder(pth) {
					watcherLog("Ignoring %s", pth)
					return filepath.SkipDir
				}

				fp := path.Join(tp, pth)
				if inArray(wfolders, fp) {
					return filepath.SkipDir
				}
				wfolders = append(wfolders, fp)

				err := watcher.Add(fp)
				if err != nil {
					watcherLog("Add watch path err:%s", err)
				} else {
					watcherLog("Watching %s", pth)
				}
			}

			return err
		})
	}
}
