package term

import (
	"errors"
	"os"

	"github.com/bitfield/script"
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
	_, err := script.Exec("tput cup 0 0").Stdout()
	return err
}

func Clear() error {
	_, err := script.Exec("clear").Stdout()
	return err
}
