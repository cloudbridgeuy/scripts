package utils

import (
	"github.com/bitfield/script"
)

func FirstOrStdin(args []string) (string, error) {
	if len(args) > 0 && args[0] != "-" {
		return args[0], nil
	}

	return script.Stdin().String()
}
