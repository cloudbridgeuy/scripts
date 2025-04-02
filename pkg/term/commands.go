package term

import (
	"github.com/bitfield/script"
)

func Clear() error {
	_, err := script.Exec("clear").Stdout()
	return err
}
