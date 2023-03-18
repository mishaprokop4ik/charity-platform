package util

import "github.com/samber/lo"

func ContainsOnlyDigits(s string) bool {
	digits := []rune{
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
	}
	for i := range s {
		if !lo.Contains(digits, rune(s[i])) {
			return false
		}
	}

	return true
}
