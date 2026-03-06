package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// DevServer orchestrates the local development environment.
// It watches the /src directory and synchronously rebuilds the /dist deterministic bundle automatically.
func DevServer(args []string) error {
	fmt.Println("Triggering initial AOT build...")
	if err := BuildProject(args, true); err != nil {
		fmt.Printf("Initial build failed: %v\n", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to initialize fsnotify: %v", err)
	}
	defer watcher.Close()

	// Recursively watch src/ directory
	srcPath := "src"
	if stat, err := os.Stat(srcPath); err == nil && stat.IsDir() {
		err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to add directories to watcher: %v", err)
		}
	} else {
		return fmt.Errorf("could not find src directory. Ensure you're in an Orbis project")
	}

	_ = watcher.Add("public")
	_ = watcher.Add("orbis.config.json")

	// Start rebuild loop (blocking background routine)
	var reloadToken int64 = 0
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Rebuild natively without complex hot-replace hooks to maintain determinism guarantees
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("\n[Dev] File modified: %s\n", event.Name)
					if err := BuildProject(args, true); err != nil {
						fmt.Printf("[Dev] AOT compilation failed: %v\n", err)
					} else {
						reloadToken++
						fmt.Println("[Dev] Rebuild successful. Triggering browser reload...")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("[Dev] Watcher error: %v\n", err)
			}
		}
	}()

	// Serve the compiled app deterministically
	port := "3000"
	fmt.Printf("\nOrbis Dev Server running at http://localhost:%s (Press CTRL+C to quit)\n", port)

	mux := http.NewServeMux()

	// Special endpoint for live-reload polling
	mux.HandleFunc("/__orbis_reload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprintf(w, "%d", reloadToken)
	})

	// Serve generated dist bundle
	fs := http.FileServer(http.Dir("dist"))
	mux.Handle("/", fs)

	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		return fmt.Errorf("server failed: %v", err)
	}

	return nil
}
