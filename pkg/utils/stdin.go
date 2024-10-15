package utils

import (
	"io"
)

func FirstOrStdin(args []string, inputReader io.Reader) (string, error) {
	if len(args) > 0 && args[0] != "-" {
		return args[0], nil
	}

	inputBytes, err := io.ReadAll(inputReader)
	if err != nil {
		return "", err
	}

	return string(inputBytes), nil
}
