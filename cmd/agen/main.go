package main

import (
	"fmt"
	"os"

	"github.com/Kegian/agen/cmd/agen/gen"
	in "github.com/Kegian/agen/cmd/agen/init"
	"github.com/Kegian/agen/cmd/agen/update"
	"github.com/Kegian/agen/cmd/agen/web"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agen",
	Short: "A tool for generating openapi/ogen files",
}

func init() {
	rootCmd.AddCommand(in.InitCmd)
	rootCmd.AddCommand(gen.GenCmd)
	rootCmd.AddCommand(web.WebCmd)
	rootCmd.AddCommand(update.UpdateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
