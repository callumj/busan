package utils

import "fmt"

func BuildName(version string) string {
	return fmt.Sprintf("%s:v%s", GlobalOptions.Name, version)
}
