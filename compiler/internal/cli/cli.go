package cli

import "fmt"

// Run is the main dispatcher for CLI commands.
// It ensures only deterministic execution paths are invoked.
func Run(command string, args []string) error {
	switch command {
	case "new":
		return NewProject(args)
	case "build":
		return BuildProject(args, false)
	case "dev":
		return DevServer(args)
	case "version":
		fmt.Println("Orbis CLI v1.0.0 - Deterministic AOT compiler")
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}
