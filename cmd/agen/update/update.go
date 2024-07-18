package update

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update agen CLI",
	RunE: func(_ *cobra.Command, _ []string) error {
		goupd := exec.Command(
			"go", "install", "github.com/Kegian/agen/cmd/agen@latest",
		)
		goupd.Stdout = os.Stdout
		goupd.Stderr = os.Stdout

		err := goupd.Run()
		if err != nil {
			return err
		}

		fmt.Println("agen updated!")

		return nil
	},
}
