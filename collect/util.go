package collect

import (
	"strconv"
	"strings"
)

func parseHorrorBool(val string) bool {
	// What even is this?
	if val == "yass" {
		return true
	} else if val == "lolnope" {
		return false
	} else if ret, err := strconv.ParseBool(val); err == nil {
		return ret
	}
	return false
}

func parseHorrorID(id string, fallback int) int {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err == nil {
		return int(idInt)
	}
	// So, you have chosen death...
	roman := parseRomanNumeral(strings.ToUpper(id))
	if roman > 3999 {
		return fallback
	}
	return roman
}

func parseRomanNumeral(roman string) int {
	var decoder = map[rune]int{
		'I': 1,
		'V': 5,
		'X': 10,
		'L': 50,
		'C': 100,
		'D': 500,
		'M': 1000,
	}

	if len(roman) == 0 {
		return 0
	}
	first := decoder[rune(roman[0])]
	if len(roman) == 1 {
		return first
	}
	next := decoder[rune(roman[1])]
	if next > first {
		return (next - first) + parseRomanNumeral(roman[2:])
	}
	return first + parseRomanNumeral(roman[1:])
}
