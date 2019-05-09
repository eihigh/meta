package main

const template = `package main

import "github.com/eihigh/meta/metafile"

func main() {
	metafile.New(
		metafile.Tools(),
		metafile.Tasks(
			metafile.Task("test", test),
		),
	)
}

func test([]string) error {
	panic("no test specified.")
}
`
