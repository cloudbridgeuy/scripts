package term

import (
	"github.com/bitfield/script"
)

func CenterCursor() error {
	_, err := script.Exec("tput cup 0 0").Stdout()
	return err
}

func Clear() error {
	_, err := script.Exec("clear").Stdout()
	return err
}
