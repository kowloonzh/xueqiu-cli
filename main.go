package main

import (
	"fmt"
	"os"

	"github.com/kowloonzh/xueqiu-cli/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand(os.Stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
