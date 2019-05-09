package metafile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	pwd, projPath string
)

func New(fns ...OptionFn) {

	// apply functional options
	var o option
	for _, fn := range fns {
		fn(&o)
	}

	// take the args given from global meta
	var xs []string
	pwd, xs = shift(os.Args)
	projPath, xs = shift(xs)

	// run commands
	var err error
	switch x, xs := shift(xs); x {
	case "setup":
		err = o.setup()
	case "tools":
		err = o.showTools()
	case "tool", "use":
		err = o.runTool(shift(xs))
	case "task":
		err = o.runTask(shift(xs))
	default:
		err = o.runTask(x, xs)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type option struct {
	tools       []string
	tasks       []task
	beforeSetup func() error
	afterSetup  func() error
}

type task struct {
	name   string
	action func(args []string) error
}

func (o *option) setup() error {

	if o.beforeSetup != nil {
		if err := o.beforeSetup(); err != nil {
			return err
		}
	}

	os.Setenv("GOBIN", toolsPath())
	for _, tool := range o.tools {
		if err := RunV("go", "install", tool); err != nil {
			return err
		}
	}

	if o.afterSetup != nil {
		if err := o.afterSetup(); err != nil {
			return err
		}
	}

	return nil
}

func (o *option) showTools() error {
	paths, _ := filepath.Glob(filepath.Join(toolsPath(), "*"))
	for _, path := range paths {
		rel, _ := filepath.Rel(toolsPath(), path)
		fmt.Println(rel)
	}
	return nil
}

func (o *option) runTool(tool string, args []string) error {
	path := filepath.Join(toolsPath(), tool)
	if isFile(path) {
		return RunV(path, args...)
	}
	return fmt.Errorf("'%v': no such tool installed", tool)
}

func (o *option) runTask(task string, args []string) error {
	for _, t := range o.tasks {
		if t.name == task {
			return t.action(args)
		}
	}
	return fmt.Errorf("'%v': undefined task", task)
}

type OptionFn func(*option)

func Tools(tools ...string) OptionFn {
	return func(o *option) {
		o.tools = tools
	}
}

type TaskFn func() task

func Task(name string, action func([]string) error) TaskFn {
	return func() task {
		return task{name: name, action: action}
	}
}

func Tasks(fns ...TaskFn) OptionFn {
	return func(o *option) {
		o.tasks = []task{}
		for _, fn := range fns {
			o.tasks = append(o.tasks, fn())
		}
	}
}

func BeforeSetup(fn func() error) OptionFn {
	return func(o *option) {
		o.beforeSetup = fn
	}
}

func AfterSetup(fn func() error) OptionFn {
	return func(o *option) {
		o.afterSetup = fn
	}
}

func toolsPath() string {
	return filepath.Join(projPath, "meta/tools")
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

func shift(xs []string) (string, []string) {
	if len(xs) == 0 {
		return "", xs
	}
	return xs[0], xs[1:]
}

func RunV(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func RunVIn(dir, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
