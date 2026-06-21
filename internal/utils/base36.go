package utils

import "strconv"

func ToBase36(num int64, length int) string {
	s := strconv.FormatInt(num, 36)
	for len(s) < length {
		s = "0" + s
	}
	return s
}
