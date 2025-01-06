package output

import (
	"fmt"
)

func Printf(format string, a ...any) (n int, err error) {
	return fmt.Printf(format, a...)
}
