package gin

import (
	"strings"
)

func inSlice(search string, array []string) bool {
	for _, s := range array {
		if strings.ToLower(search) == strings.ToLower(s) {
			return true
		}
	}
	return false
}
