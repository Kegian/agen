package init

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init <project_name>",
	Args:  cobra.MinimumNArgs(1),
	Short: "Init new project in current folder",
	RunE: func(_ *cobra.Command, args []string) error {
		_, err := exec.LookPath("gonew")
		if err != nil {
			return errors.New(`gonew not found, install it via "go install golang.org/x/tools/cmd/gonew@latest"`)
		}

		gonew := exec.Command(
			"gonew",
			"github.com/Kegian/agen/examples/server",
			args[0],
			".",
		)
		gonew.Stdout = os.Stdout
		gonew.Stderr = os.Stdout

		err = gonew.Run()
		if err != nil {
			return err
		}

		gotidy := exec.Command(
			"go", "mod", "tidy",
		)
		gotidy.Stdout = os.Stdout
		gotidy.Stderr = os.Stdout

		err = gotidy.Run()
		if err != nil {
			return err
		}

		fmt.Println("Project successfully initialized!")
		fmt.Println("  make deps -- install required dependencies")
		fmt.Println("  make gen  -- generate api/sql go files")
		fmt.Println("  make run  -- run project")

		return nil
	},
}
