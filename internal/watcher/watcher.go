package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/polygone-app/kustomize-watcher/internal/applier"
)

type Watcher struct {
	glob    string
	applier *applier.Applier
	logger  *slog.Logger
	fw      *fsnotify.Watcher
	mu      sync.Mutex
	timers  map[string]*time.Timer
}

func New(glob string, a *applier.Applier, logger *slog.Logger) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify: %w", err)
	}

	w := &Watcher{
		glob:    glob,
		applier: a,
		logger:  logger,
		fw:      fw,
		timers:  make(map[string]*time.Timer),
	}

	if err := w.discoverAndWatch(); err != nil {
		fw.Close()
		return nil, err
	}

	for _, dir := range fw.WatchList() {
		logger.Info("initial apply", "dir", dir)
		if err := a.Apply(context.Background(), dir); err != nil {
			logger.Error("initial apply failed", "dir", dir, "err", err)
		}
	}

	return w, nil
}

func (w *Watcher) discoverAndWatch() error {
	matches, err := doublestar.FilepathGlob(w.glob)
	if err != nil {
		return fmt.Errorf("glob %q: %w", w.glob, err)
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || !info.IsDir() {
			continue
		}
		if !hasKustomization(match) {
			continue
		}
		if err := w.fw.Add(match); err != nil {
			w.logger.Warn("cannot watch dir", "path", match, "err", err)
			continue
		}
		w.logger.Info("watching", "path", match)
	}
	return nil
}

func hasKustomization(dir string) bool {
	for _, name := range []string{"kustomization.yaml", "kustomization.yml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func (w *Watcher) Run(ctx context.Context) error {
	defer w.fw.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-w.fw.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) ||
				event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				dir := filepath.Dir(event.Name)
				w.scheduleApply(ctx, dir)
			}
		case err, ok := <-w.fw.Errors:
			if !ok {
				return nil
			}
			w.logger.Error("watcher error", "err", err)
		}
	}
}

func (w *Watcher) scheduleApply(ctx context.Context, dir string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, ok := w.timers[dir]; ok {
		t.Stop()
	}
	w.timers[dir] = time.AfterFunc(500*time.Millisecond, func() {
		w.logger.Info("applying", "dir", dir)
		if err := w.applier.Apply(ctx, dir); err != nil {
			w.logger.Error("apply failed", "dir", dir, "err", err)
		}
	})
}
