package scanner

import (
	"bytes"
	"errors"
	"strconv"
	"unicode"

	"github.com/harshagw/viri/internal/token"
)

type Scanner struct {
	source  *bytes.Buffer
	current int
	start   int
	line    int
	tokens  []token.Token
}

func New(source *bytes.Buffer) *Scanner {
	return &Scanner{
		source:  source,
		current: 0,
		start:   0,
		line:    1,
		tokens:  []token.Token{},
	}
}

func (s *Scanner) Scan() ([]token.Token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		if err := s.scanToken(); err != nil {
			return nil, err
		}
	}

	s.start = s.current
	s.addToken(token.EOF)
	return s.tokens, nil
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= s.source.Len()
}

func (s *Scanner) scanToken() error {
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
		if err := s.scanString(); err != nil {
			return err
		}
	default:
		if unicode.IsDigit(rune(c)) {
			s.scanNumber()
		} else if unicode.IsLetter(rune(c)) {
			s.scanIdentifier()
		} else {
			return errors.New("unexpected character: " + string(c))
		}
	}
	return nil
}

// Returns the current character and advances the current pointer.
func (s *Scanner) advance() byte {
	c := s.source.Bytes()[s.current]
	s.current++
	return c
}

// Matches the current character with the expected character and then advances the pointer if it matches.
func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}
	if s.source.Bytes()[s.current] != expected {
		return false
	}
	s.current++
	return true
}

// Returns the current character without advancing the pointer.
func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source.Bytes()[s.current]
}

// Returns the character after the current one without advancing.
func (s *Scanner) peekNext() byte {
	if s.current+1 >= s.source.Len() {
		return '\000'
	}
	return s.source.Bytes()[s.current+1]
}

func (s *Scanner) scanString() error {
	startLine := s.line

	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		return errors.New("unterminated string start at line: " + strconv.Itoa(startLine))
	}

	// The closing quote
	s.advance()

	// Trim the surrounding quotes
	value := string(s.source.Bytes()[s.start+1 : s.current-1])
	s.addTokenWithLiteral(token.STRING, value)
	return nil
}

func (s *Scanner) scanNumber() {
	for unicode.IsDigit(rune(s.peek())) {
		s.advance()
	}

	// Look for a fractional part
	if s.peek() == '.' && unicode.IsDigit(rune(s.peekNext())) {
		s.advance()

		for unicode.IsDigit(rune(s.peek())) {
			s.advance()
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
	for unicode.IsLetter(rune(s.peek())) || unicode.IsDigit(rune(s.peek())) {
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
	s.tokens = append(s.tokens, token.New(tokenType, text, literal, s.line))
}

// Returns the string starting from start to current.
func (s *Scanner) getLexeme() string {
	buf := s.source.Bytes()
	if s.start < 0 || s.current > len(buf) || s.start > s.current {
		return ""
	}
	return string(buf[s.start:s.current])
}
