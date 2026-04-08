package tools

import "fmt"

func ErrToolNotFound(name string) error {
	return fmt.Errorf("tool %q not found", name)
}
