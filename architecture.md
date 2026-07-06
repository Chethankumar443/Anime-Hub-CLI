```
       +---------------------------------------------+
       |             Bubbletea TUI View              |
       +----------------------+----------------------+
                              |
                     Sends IO State Cmds
                              v
       +---------------------------------------------+
       |          TUI Engine / State Machine         |
       +----------------------+----------------------+
                              |
               Orchestrates Calls via Interface
                              v
       +---------------------------------------------+
       |               FallbackManager               |
       +----------------------+----------------------+
                              |
                     Loops Through Strategies
                              v
         +--------------------+--------------------+
         |                                         |
         v                                         v
+------------------+                      +------------------+
|  Consumet Client |                      | Alternative Prov |
+--------+---------+                      +--------+---------+
         |                                         |
         +--------------------+--------------------+
                              |
                       Routes JSON HTTP
                              v
       +---------------------------------------------+
       |   Embedded Background Provider Lifecycle    |
       +---------------------------------------------+

```

### Decoupled Data Extraction

The scraper layer is deliberately decoupled from the TUI core. Writing native regex or DOM scrapers directly into Go introduces excessive maintenance overhead. Instead, the application treats the scraper as an abstract, independent web service, protecting the application's runtime code from upstream changes.

### Strategy Pattern for Providers

A unified Go interface governs data retrieval. The system routes calls through an orchestration module that guarantees structural safety even if a provider goes completely offline during a session.

```go
package main

import (
	"errors"
	"log"
)

type Anime struct {
	ID    string
	Title string
}

type Episode struct {
	ID     string
	Number int
}

// AnimeProvider abstracts individual content scraper implementations.
type AnimeProvider interface {
	Search(query string) ([]Anime, error)
	GetEpisodes(animeID string) ([]Episode, error)
	GetStreamURL(episodeID string, lang string) (string, error)
}

// FallbackManager orchestrates sequential provider evaluation.
type FallbackManager struct {
	providers []AnimeProvider
}

func (fm *FallbackManager) GetStreamURL(episodeID, lang string) (string, error) {
	for _, p := range fm.providers {
		url, err := p.GetStreamURL(episodeID, lang)
		if err == nil {
			return url, nil 
		}
		log.Printf("Primary provider failure: %v. Escalating to alternative route...", err)
	}
	return "", errors.New("exhausted all available provider routes without resolving stream")
}

```

### Local Runtime Lifecycle Management (No-Docker Architecture)

To eliminate heavy end-user system configurations, the application avoids local Docker setups. Instead, it compiles the provider code into a self-contained background binary utilizing packaging utilities (`pkg`).

The Go runtime supervises this background binary's lifecycle, managing port routing and system termination sequences automatically.

```go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

type ConsumetManager struct {
	cmd                  *exec.Cmd
	port                 string
	pathToConsumetBinary string
}

func (cm *ConsumetManager) Start() error {
	// Dynamically scan for open ports if default is bound
	if cm.isPortInUse(cm.port) {
		return nil 
	}

	cm.cmd = exec.Command(cm.pathToConsumetBinary)
	cm.cmd.Env = append(os.Environ(), "PORT="+cm.port)

	// Detach child processes and mask consoles under Windows platforms
	if runtime.GOOS == "windows" {
		cm.cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	if err := cm.cmd.Start(); err != nil {
		return fmt.Errorf("initialization failure on provider binary: %w", err)
	}

	return cm.waitForReady()
}

func (cm *ConsumetManager) waitForReady() error {
	url := fmt.Sprintf("http://localhost:%s/health", cm.port)
	for i := 0; i < 30; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("timeout limit reached waiting for child service initialization")
}

func (cm *ConsumetManager) isPortInUse(port string) bool {
	ln, err := http.Get("http://localhost:" + port + "/health")
	if err == nil {
		ln.Body.Close()
		return true
	}
	return false
}

func (cm *ConsumetManager) Stop() {
	if cm.cmd != nil && cm.cmd.Process != nil {
		_ = cm.cmd.Process.Kill()
	}
}

```
