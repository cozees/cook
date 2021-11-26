package parser

func lower(ch rune) rune { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isLetter(ch rune) bool { return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' }

func isSpace(r rune) bool { return r == ' ' || r == '\t' }

func isLineFeed(r rune) bool { return r == '\n' || r == '\r' }

func isNotLineFeed(r rune) bool { return r != '\n' && r != '\r' }

func isArgumentStringTerminate(r rune) bool {
	switch r {
	case '>', '<', '\\', ' ', '\t', '\n', '\r', eof:
		return true
	default:
		return false
	}
}
