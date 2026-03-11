package term

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

func GetColumns() (int, error) {
	if !IsTerm() {
		return -1, errors.New("not a terminal")
	}

	w, _, err := term.GetSize(0)

	if err != nil {
		return -1, err
	}

	return w, nil
}

func IsTerm() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func CenterCursor() error {
	_, err := fmt.Fprint(os.Stdout, "\033[H")
	return err
}

func Clear() error {
	_, err := fmt.Fprint(os.Stdout, "\033[2J\033[H")
	return err
}

func ClearFromCursor() error {
	_, err := fmt.Fprint(os.Stdout, "\033[J")
	return err
}
