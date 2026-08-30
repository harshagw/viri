package scanner

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/harshagw/viri/internal/interpreter/token"
)

type Scanner struct {
	source   *bytes.Buffer
	current  int
	start    int
	line     int
	tokens   []token.Token
	filePath *string
	errs     []string
}

func New(source *bytes.Buffer, filePath *string) *Scanner {
	return &Scanner{
		source:   source,
		current:  0,
		start:    0,
		line:     1,
		tokens:   []token.Token{},
		filePath: filePath,
	}
}

// Scan tokenizes the whole source. Errors are collected with their line
// numbers and scanning continues past them, so one bad character reports
// every problem in the file instead of aborting at the first.
func (s *Scanner) Scan() ([]token.Token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.start = s.current
	s.addToken(token.EOF)

	if len(s.errs) > 0 {
		return s.tokens, errors.New(strings.Join(s.errs, "; "))
	}
	return s.tokens, nil
}

func (s *Scanner) error(message string) {
	s.errs = append(s.errs, fmt.Sprintf("line %d: %s", s.line, message))
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= s.source.Len()
}

func (s *Scanner) scanToken() {
	c := s.advance()

	switch c {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case '[':
		s.addToken(token.LEFT_BRACKET)
	case ']':
		s.addToken(token.RIGHT_BRACKET)
	case ',':
		s.addToken(token.COMMA)
	case ':':
		s.addToken(token.COLON)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '*':
		s.addToken(token.STAR)
	case '!':
		if s.match('=') {
			s.addToken(token.BANG_EQUAL)
		} else {
			s.addToken(token.BANG)
		}
	case '=':
		if s.match('=') {
			s.addToken(token.EQUAL_EQUAL)
		} else {
			s.addToken(token.EQUAL)
		}
	case '<':
		if s.match('=') {
			s.addToken(token.LESS_EQUAL)
		} else {
			s.addToken(token.LESS)
		}
	case '>':
		if s.match('=') {
			s.addToken(token.GREATER_EQUAL)
		} else {
			s.addToken(token.GREATER)
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(token.SLASH)
		}
	case '\t', '\r', ' ':
	case '\n':
		s.line++
	case '"':
		s.scanString()
	default:
		if unicode.IsDigit(c) {
			s.scanNumber()
		} else if isIdentifierStart(c) {
			s.scanIdentifier()
		} else {
			s.error("unexpected character: " + string(c))
		}
	}
}

func isIdentifierStart(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}

func isIdentifierPart(c rune) bool {
	return c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

// Returns the current rune and advances past it (UTF-8 aware).
func (s *Scanner) advance() rune {
	r, size := utf8.DecodeRune(s.source.Bytes()[s.current:])
	if r == utf8.RuneError && size <= 1 {
		// Invalid byte: consume it so scanning can continue
		s.current++
		return utf8.RuneError
	}
	s.current += size
	return r
}

// Matches the current rune with the expected rune and then advances the pointer if it matches.
func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if s.peek() != expected {
		return false
	}
	s.advance()
	return true
}

// Returns the current rune without advancing the pointer.
func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}
	r, _ := utf8.DecodeRune(s.source.Bytes()[s.current:])
	return r
}

// Returns the rune after the current one without advancing.
func (s *Scanner) peekNext() rune {
	if s.isAtEnd() {
		return 0
	}
	_, size := utf8.DecodeRune(s.source.Bytes()[s.current:])
	if s.current+size >= s.source.Len() {
		return 0
	}
	r, _ := utf8.DecodeRune(s.source.Bytes()[s.current+size:])
	return r
}

func (s *Scanner) scanString() {
	startLine := s.line
	var value strings.Builder

	for s.peek() != '"' && !s.isAtEnd() {
		c := s.advance()
		switch c {
		case '\n':
			s.line++
			value.WriteRune(c)
		case '\\':
			if s.isAtEnd() {
				break
			}
			escape := s.advance()
			switch escape {
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case 'r':
				value.WriteByte('\r')
			case '"':
				value.WriteByte('"')
			case '\\':
				value.WriteByte('\\')
			case '0':
				value.WriteByte(0)
			default:
				s.error("invalid escape sequence '\\" + string(escape) + "'")
			}
		default:
			value.WriteRune(c)
		}
	}

	if s.isAtEnd() {
		s.error("unterminated string starting at line " + strconv.Itoa(startLine))
		return
	}

	// The closing quote
	s.advance()

	s.addTokenWithLiteral(token.STRING, value.String())
}

func (s *Scanner) scanNumber() {
	for unicode.IsDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part
	if s.peek() == '.' && unicode.IsDigit(s.peekNext()) {
		s.advance()

		for unicode.IsDigit(s.peek()) {
			s.advance()
		}
	}

	// Look for an exponent: 1e3, 2.5E-4, 1e+10
	if s.peek() == 'e' || s.peek() == 'E' {
		next := s.peekNext()
		if unicode.IsDigit(next) || next == '+' || next == '-' {
			s.advance() // e / E
			if s.peek() == '+' || s.peek() == '-' {
				s.advance()
			}
			if !unicode.IsDigit(s.peek()) {
				s.error("exponent has no digits")
			}
			for unicode.IsDigit(s.peek()) {
				s.advance()
			}
		}
	}

	text := s.getLexeme()
	var value interface{}
	if len(text) > 0 {
		if num, err := strconv.ParseFloat(text, 64); err == nil {
			value = num
		}
	}
	s.addTokenWithLiteral(token.NUMBER, value)
}

func (s *Scanner) scanIdentifier() {
	for isIdentifierPart(s.peek()) {
		s.advance()
	}

	text := s.getLexeme()
	tokenType := token.LookupKeyword(text)
	if tokenType == token.TRUE {
		s.addTokenWithLiteral(token.TRUE, true)
	} else if tokenType == token.FALSE {
		s.addTokenWithLiteral(token.FALSE, false)
	} else if tokenType == token.IDENTIFIER {
		s.addTokenWithLiteral(token.IDENTIFIER, text)
	} else {
		s.addToken(tokenType)
	}
}

func (s *Scanner) addToken(tokenType token.Type) {
	s.addTokenWithLiteral(tokenType, nil)
}

func (s *Scanner) addTokenWithLiteral(tokenType token.Type, literal interface{}) {
	text := s.getLexeme()
	tok := token.New(tokenType, text, literal, s.line, s.filePath)
	s.tokens = append(s.tokens, tok)
}

// Returns the string starting from start to current.
func (s *Scanner) getLexeme() string {
	buf := s.source.Bytes()
	if s.start < 0 || s.current > len(buf) || s.start > s.current {
		return ""
	}
	return string(buf[s.start:s.current])
}
