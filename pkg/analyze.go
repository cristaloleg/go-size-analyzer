package pkg

import (
	"github.com/goretk/gore"
)

func Analyze(path string) error {
	file, err := gore.Open(path)
	if err != nil {
		return err
	}

	_, err = Collect(file)

	return err
}
