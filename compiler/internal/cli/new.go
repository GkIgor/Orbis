package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewProject generates the official Orbis project layout deterministically.
func NewProject(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing project name. Usage: orbis new <project-name>")
	}

	projectName := args[0]
	fmt.Printf("Creating new Orbis project structure: %s\n", projectName)

	// Explicit directory structure required by the deterministic compilation
	dirs := []string{
		projectName,
		filepath.Join(projectName, "src", "components"),
		filepath.Join(projectName, "src", "states"),
		filepath.Join(projectName, "public"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Mandatory configuration files and default component hooks
	files := map[string]string{
		filepath.Join(projectName, "src", "app.component.ts"):   appComponentTS,
		filepath.Join(projectName, "src", "app.component.html"): appComponentHTML,
		filepath.Join(projectName, "src", "app.component.scss"): "",
		filepath.Join(projectName, "public", "index.html"):      indexHTML(projectName),
		filepath.Join(projectName, "orbis.config.json"):         orbisConfigJSON,
		filepath.Join(projectName, "package.json"):              packageJSON(projectName),
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", path, err)
		}
	}

	fmt.Println("Project scaffold created successfully.")
	return nil
}

// Ensure the default App Component explicitly conforms to exactly what was requested.
// Specifically keeping logic out of the template per DSL rules.
const appComponentTS = `@Component({
  selector: "app-root"
})
export class AppComponent {

  title = "Orbis App"

  onInit() {
    this.render()
  }

}
`

const appComponentHTML = `<div class="app">
  <h1>{{ title }}</h1>
</div>
`

func indexHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body>
    <div id="app"></div>
    <script type="module" src="/bundle.js"></script>
</body>
</html>
`, title)
}

const orbisConfigJSON = `{
  "entry": "src/app.component.ts",
  "outDir": "dist",
  "styles": [
    "src/app.component.scss"
  ]
}
`

func packageJSON(name string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Orbis project scaffold",
  "type": "module",
  "scripts": {
    "dev": "orbis dev",
    "build": "orbis build"
  },
  "dependencies": {
    "@orbisui/runtime": "latest"
  }
}
`, name)
}
