package legacy

import (
	"strconv"
)

func isStringInIntSlice(nsp string, nps []int) bool {
	nid, _ := strconv.Atoi(nsp)

	for _, nsv := range nps {
		if nid == nsv {
			return true
		}
	}

	return false
}
