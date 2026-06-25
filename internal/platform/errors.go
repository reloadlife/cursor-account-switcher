package platform

import "fmt"

func fmtError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
