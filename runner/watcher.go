package runner

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && isWatchedFile(event.Name) {
					//如果修改时间是30秒之前，就忽略，多数出现在windows系统，多次触发Write事件，其实文件并没有修改
					if runtime.GOOS == "windows" {
						if stat, err := os.Stat(event.Name); err != nil {
							watcherLog("error: %s", err)
							return
						} else {
							if stat.ModTime().Add(time.Second * 10).Before(time.Now()) {
								watcherLog("sending event %s, [ignore]", event)
								return
							}
						}
					}

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
			if err != nil {
				watcherLog("path: %#q; info: %v; error: %v\n", pth, info, err)
				return err
			}
			if info.IsDir() && !isTmpDir(pth) {
				if len(pth) > 1 && strings.HasPrefix(filepath.Base(pth), ".") {
					return filepath.SkipDir
				}

				if isIgnoredFolder(pth) {
					watcherLog("Ignoring %#q", pth)
					return filepath.SkipDir
				}

				apath, err := filepath.Abs(pth)
				if err != nil {
					watcherLog("Add watch path error. path:%#q, err:%s", pth, err)
					return filepath.SkipDir
				}
				if inArray(wfolders, apath) {
					return filepath.SkipDir
				}
				wfolders = append(wfolders, apath)

				err = watcher.Add(apath)
				if err != nil {
					watcherLog("Add watch path error. path:%s, err:%s", pth, err)
				} else {
					watcherLog("Watching %#q", pth)
				}
			}

			return err
		})
	}
}
