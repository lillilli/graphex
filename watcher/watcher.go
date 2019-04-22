package watcher

import (
	"context"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/lillilli/logger"
)

type watcher struct {
	dir    string
	files  map[string]bool
	events chan *Event
	log    logger.Logger
	sync.RWMutex
}

type Watcher interface {
	Start(ctx context.Context) error
	UpdatesChannel() <-chan *Event

	State() []string
	FileState(name string) (*FileData, error)
}

func New(dir string) Watcher {
	return &watcher{
		dir:    dir,
		files:  make(map[string]bool),
		events: make(chan *Event),
		log:    logger.NewLogger("watcher"),
	}
}

func (w *watcher) Start(ctx context.Context) error {
	if err := w.updateFilesCache(); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go w.startWatch(ctx, watcher)
	return watcher.Add(w.dir)
}

func (w *watcher) startWatch(ctx context.Context, watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			fileName := strings.TrimPrefix(event.Name, w.dir+"/")

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				w.log.Debugf("File %q renamed", fileName)
				delete(w.files, fileName)
				w.events <- &Event{Type: RemoveState, Name: fileName}
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				w.log.Debugf("File %q modified", fileName)
				go w.handleFileModify(event.Name, fileName, ModifyState)
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				w.log.Debugf("File %q created", fileName)
				w.files[fileName] = true
				go w.handleFileModify(event.Name, fileName, CreateState)
			}

		case err := <-watcher.Errors:
			w.log.Errorf("Watcher return error: %v", err)
			watcher.Close()
			return

		case <-ctx.Done():
			watcher.Close()
			return
		}
	}
}

func (w *watcher) updateFilesCache() error {
	fileInfos, err := ioutil.ReadDir(w.dir)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()

		if fileInfo.IsDir() || !strings.HasSuffix(fileName, ".txt") {
			continue
		}

		w.files[fileName] = true
	}

	return nil
}

func (w *watcher) handleFileModify(fullPath, name, modifyType string) {
	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		w.log.Errorf("Reading file failed: %v", err)
		return
	}

	w.events <- &Event{Type: modifyType, Name: name, Values: parseFile(b)}
}

func (w *watcher) State() []string {
	files := make([]string, 0)
	w.RLock()

	for fileName := range w.files {
		files = append(files, fileName)
	}

	w.RUnlock()
	return files
}

func (w *watcher) FileState(name string) (*FileData, error) {
	b, err := ioutil.ReadFile(w.dir + "/" + name)
	return parseFile(b), err
}

func (w *watcher) UpdatesChannel() <-chan *Event {
	return w.events
}
