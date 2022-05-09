package scanner

func lower(ch rune) rune { return ('a' - 'A') | ch }

func isPunctuation(ch rune) bool {
	switch ch {
	case '=', '(', ')', '[', ']', '{', '}', ',', ';', ':':
		return true
	}
	return false
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}
