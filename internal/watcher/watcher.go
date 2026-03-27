package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/rushteam/skills-cli/internal/agent"
	"github.com/rushteam/skills-cli/internal/config"
	"github.com/rushteam/skills-cli/internal/skill"
)

const debounceDelay = 300 * time.Millisecond

var (
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type Watcher struct {
	cfg       *config.Config
	direction string
	watcher   *fsnotify.Watcher
	stopCh    chan struct{}
	mu        sync.Mutex
	pending   map[string]time.Time
}

func New(cfg *config.Config, direction string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		cfg:       cfg,
		direction: direction,
		watcher:   w,
		stopCh:    make(chan struct{}),
		pending:   make(map[string]time.Time),
	}, nil
}

func (w *Watcher) Start() error {
	dirs := w.collectWatchDirs()
	if len(dirs) == 0 {
		return fmt.Errorf("no directories to watch")
	}

	for _, dir := range dirs {
		if err := w.addRecursive(dir); err != nil {
			slog.Warn("failed to watch directory", "dir", dir, "error", err)
		}
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("Watching %d directories (direction: %s)", len(dirs), w.direction)))
	for _, d := range dirs {
		fmt.Println(dimStyle.Render(fmt.Sprintf("  %s", agent.ShortenPath(d))))
	}

	go w.eventLoop()
	return nil
}

func (w *Watcher) Stop() {
	close(w.stopCh)
	w.watcher.Close()
}

func (w *Watcher) Wait() {
	<-w.stopCh
}

func (w *Watcher) collectWatchDirs() []string {
	var dirs []string
	centralDir := config.SkillsHome()

	switch w.direction {
	case config.WatchCentralToAgents:
		dirs = append(dirs, centralDir)
	case config.WatchAgentsToCentral:
		dirs = append(dirs, w.agentDirs()...)
	case config.WatchBidirectional:
		dirs = append(dirs, centralDir)
		dirs = append(dirs, w.agentDirs()...)
	}

	var existing []string
	for _, d := range dirs {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			existing = append(existing, d)
		}
	}
	return existing
}

func (w *Watcher) agentDirs() []string {
	var dirs []string
	seen := make(map[string]bool)

	for _, ag := range w.cfg.Agents {
		dir := config.ResolveGlobalPath(ag)
		if !seen[dir] {
			dirs = append(dirs, dir)
			seen[dir] = true
		}
	}

	for _, proj := range w.cfg.Projects {
		agents := proj.Agents
		if len(agents) == 0 {
			agents = agent.DetectProjectAgents(proj.Path, w.cfg.Agents)
		}
		for _, agName := range agents {
			dir := agent.ResolveProjectSkillsDir(proj.Path, agName, w.cfg.Agents)
			if dir != "" && !seen[dir] {
				dirs = append(dirs, dir)
				seen[dir] = true
			}
		}
	}
	return dirs
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return w.watcher.Add(path)
		}
		return nil
	})
}

func (w *Watcher) eventLoop() {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-w.stopCh:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.mu.Lock()
				w.pending[event.Name] = time.Now()
				w.mu.Unlock()
				timer.Reset(debounceDelay)
			}
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.watcher.Add(event.Name)
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "error", err)
		case <-timer.C:
			w.processBatch()
		}
	}
}

func (w *Watcher) processBatch() {
	w.mu.Lock()
	changes := make(map[string]time.Time, len(w.pending))
	for k, v := range w.pending {
		changes[k] = v
	}
	w.pending = make(map[string]time.Time)
	w.mu.Unlock()

	if len(changes) == 0 {
		return
	}

	centralDir := config.SkillsHome()
	fmt.Println(infoStyle.Render(fmt.Sprintf("\n[%s] Detected %d change(s), syncing...", time.Now().Format("15:04:05"), len(changes))))

	isCentralChange := false
	for path := range changes {
		rel, err := filepath.Rel(centralDir, path)
		if err == nil && !filepath.IsAbs(rel) && rel[0] != '.' {
			isCentralChange = true
			break
		}
	}

	switch w.direction {
	case config.WatchCentralToAgents:
		w.pushAll()
	case config.WatchAgentsToCentral:
		w.pullAll()
	case config.WatchBidirectional:
		if isCentralChange {
			w.pushAll()
		} else {
			w.pullAll()
		}
	}
}

func (w *Watcher) pushAll() {
	centralDir := config.SkillsHome()
	skillNames, _ := agent.ScanSkillsInDir(centralDir)

	for _, ag := range w.cfg.Agents {
		dir := config.ResolveGlobalPath(ag)
		w.syncSkillsTo(centralDir, dir, skillNames)
	}
	for _, proj := range w.cfg.Projects {
		agents := proj.Agents
		if len(agents) == 0 {
			agents = agent.DetectProjectAgents(proj.Path, w.cfg.Agents)
		}
		for _, agName := range agents {
			dir := agent.ResolveProjectSkillsDir(proj.Path, agName, w.cfg.Agents)
			if dir != "" {
				w.syncSkillsTo(centralDir, dir, skillNames)
			}
		}
	}
}

func (w *Watcher) pullAll() {
	centralDir := config.SkillsHome()
	allDirs := w.agentDirs()
	for _, dir := range allDirs {
		skillNames, _ := agent.ScanSkillsInDir(dir)
		w.syncSkillsTo(dir, centralDir, skillNames)
	}
}

func (w *Watcher) syncSkillsTo(srcBase, dstBase string, skillNames []string) {
	os.MkdirAll(dstBase, 0o755)
	for _, name := range skillNames {
		srcDir := filepath.Join(srcBase, name)
		dstDir := filepath.Join(dstBase, name)

		hasDiff, _ := skill.HasDifferences(srcDir, dstDir)
		if !hasDiff {
			continue
		}

		os.RemoveAll(dstDir)
		if err := skill.CopyDir(srcDir, dstDir); err != nil {
			fmt.Println(errStyle.Render(fmt.Sprintf("  sync %s failed: %v", name, err)))
			continue
		}
		fmt.Println(okStyle.Render(fmt.Sprintf("  synced: %s -> %s", name, agent.ShortenPath(dstBase))))
	}
}
