package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli"
)

var (
	projPath string
)

func main() {

	// find metafile.go above from the current directory.
	path, _ := filepath.Abs(".")
	for !(path == "." || path == "/") {
		if isFile(filepath.Join(path, "metafile.go")) {
			projPath = path
			break
		}
		path = filepath.Dir(path)
	}

	// if metafile.go not found, run initialization
	if projPath == "" {
		beforeInit()
		return
	}

	// if projPath specified, go.mod required in the same directory.
	if !isFile(filepath.Join(projPath, "go.mod")) {
		fmt.Printf("go.mod required in '%v'", projPath)
		os.Exit(1)
	}

	// ensure directories
	for _, dir := range []string{
		metaPath(),
		toolsPath(),
	} {
		os.MkdirAll(dir, 0777)
	}

	app := cli.NewApp()
	// TODO configure app using information from local meta app.
	app.Action = func(c *cli.Context) error {
		if err := cmdInProj("go", "build", "-o", filepath.Clean("meta/bin")).Run(); err != nil {
			return err
		}
		if c.NArg() == 0 {
			cli.ShowAppHelp(c)
			return nil
		}
		if c.NArg() == 1 {
			switch c.Args()[0] {
			case "tool", "use", "task":
				cli.ShowAppHelp(c)
				return nil
			}
		}
		args := append([]string{projPath}, []string(c.Args())...)
		cmd := cmdInProj(binPath(), args...)

		// when `meta use`, inherits the current directory.
		if c.Args()[0] == "use" {
			cmd.Dir = ""
		}

		// otherwise cmd is launched from project root directory.
		return cmd.Run()
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cmdInProj(cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Dir = projPath
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}

// beforeInit runs a cli only having a subcommand to initialize.
func beforeInit() {
	app := cli.NewApp()
	app.Description = "Please run `meta init` to create initial metafile.go."
	app.Commands = []cli.Command{
		{Name: "init", Action: initialize},
	}
	app.Run(os.Args)
}

func initialize(*cli.Context) error {
	if err := ioutil.WriteFile("metafile.go", []byte(template), 0666); err != nil {
		return err
	}
	fmt.Println("metafile.go created.")
	fmt.Println("Make sure go.mod exists.")
	return nil
}

func metaPath() string {
	return filepath.Join(projPath, "meta")
}

func binPath() string {
	return filepath.Join(metaPath(), "bin")
}

func toolsPath() string {
	return filepath.Join(metaPath(), "tools")
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
