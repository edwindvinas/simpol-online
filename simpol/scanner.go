//Scanner - scans the codes
package simpol

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"fmt"
	"net/http"
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
	hw http.ResponseWriter
	hr *http.Request
	isDebug string
}

// NewScanner returns a new instance of Scanner.
func NewScanner(w http.ResponseWriter, r *http.Request, isDebug string, b io.Reader) *Scanner {
	if isDebug == "true" {
	fmt.Fprintf(w, "[X0001] NewScanner()\n")
	}
	return &Scanner{r: bufio.NewReader(b), hw: w, hr: r, isDebug: isDebug}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0002] SCANNER: s.Scan()\n")
	}
	// Read the next rune.
	ch := s.read()
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0003] SCANNER: %v\n", string(ch))
	}

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	// If we see a digit then consume as a number.
	if isWhitespace(ch) {
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0004] SCANNER: isWhitespace: %v\n", string(ch))
		}
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0005] SCANNER: isLetter: %v\n", string(ch))
		}
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0006] SCANNER: isDigit: %v\n", string(ch))
		}
		s.unread()
		return s.scanIdent()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0007] SCANNER: eof: %v\n", string(ch))
		}
		return EOF, ""
	case '*':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0008] SCANNER: ASTERISK: %v\n", string(ch))
		}
		return ASTERISK, string(ch)
	case ',':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0009] SCANNER: COMMA: %v\n", string(ch))
		}
		return COMMA, string(ch)
	case '{':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0010] SCANNER: OPEN_CURLY_BRACKET: %v\n", string(ch))
		}
		return OPEN_CURLY_BRACKET, string(ch)
	case '}':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0011] SCANNER: CLOSE_CURLY_BRACKET: %v\n", string(ch))
		}
		return CLOSE_CURLY_BRACKET, string(ch)
	case '$':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0012] SCANNER: DOLLAR_SYMBOL: %v\n", string(ch))
		}
		return DOLLAR_SYMBOL, string(ch)
	case '/':
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0013] SCANNER: SLASH_SYMBOL: %v\n", string(ch))
		}
		return SLASH_SYMBOL, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0018] SCANNER: s.scanWhitespace()\n")
	}
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			if s.isDebug == "true" {
			fmt.Fprintf(s.hw, "[X0019] SCANNER: !isWhitespace: %v\n", string(ch))
			}
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
			if s.isDebug == "true" {
			fmt.Fprintf(s.hw, "[X0022] SCANNER: WriteRune: %v\n", string(ch))
			}
		}
		
	}

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0027] SCANNER: s.scanIdent\n")
	}
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			if s.isDebug == "true" {
			fmt.Fprintf(s.hw, "[X0028] SCANNER: !isLetter && !isDigit: %v\n", string(ch))
			}
			s.unread()
			break
		} else {
			if s.isDebug == "true" {
			fmt.Fprintf(s.hw, "[X0029] SCANNER: WriteRune: %v\n", string(ch))
			}
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch strings.ToUpper(buf.String()) {
	case "VARIABLE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0032] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		FL_CURR_VARS_SECTION = true
		return VARIABLE, buf.String()		
	case "STG":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0033] SCANNER: %v\n", strings.ToUpper(buf.String()))
		fmt.Fprintf(s.hw, "[X0034] SCANNER: $$$$$$$$FOUND VARIABLE DECLARATION $$$$$$ %v\n", strings.ToUpper(buf.String()))
		}
		THIS_CURR_VAR_IDENT = "STG"
		FL_SAW_VARIABLE_DECL = true
		return STG, buf.String()	
	case "INT":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0035] SCANNER: %v\n", strings.ToUpper(buf.String()))
		fmt.Fprintf(s.hw, "[X0036] SCANNER: $$$$$$$$FOUND VARIABLE DECLARATION $$$$$$ %v\n", strings.ToUpper(buf.String()))
		}
		THIS_CURR_VAR_IDENT = "INT"
		FL_SAW_VARIABLE_DECL = true
		return INT, buf.String()
	case "BLN":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0037] SCANNER: %v\n", strings.ToUpper(buf.String()))
		fmt.Fprintf(s.hw, "[X0038] SCANNER: $$$$$$$$FOUND VARIABLE DECLARATION $$$$$$ %v\n", strings.ToUpper(buf.String()))
		}
		THIS_CURR_VAR_IDENT = "BLN"
		FL_SAW_VARIABLE_DECL = true
		return BLN, buf.String()
	case "CODE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0039] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		FL_CURR_CODE_SECTION = true
		return CODE, buf.String()	
	case "PUT":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0040] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return PUT, buf.String()
	case "IN":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0041] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return IN, buf.String()
	case "ASK":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0042] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return ASK, buf.String()
	case "PRT":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0043] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return PRT, buf.String()			
	case "ADD":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0044] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return ADD, buf.String()
	case "SUB":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0045] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return SUB, buf.String()
	case "MUL":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0046] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return MUL, buf.String()
	case "DIV":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0047] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return DIV, buf.String()
	case "MOD":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0048] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return MOD, buf.String()
	case "GRT":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0049] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return GRT, buf.String()
	case "GRE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0050] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return GRE, buf.String()	
	case "LET":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0051] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return LET, buf.String()	
	case "LEE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0052] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return LEE, buf.String()	
	case "EQL":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0053] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return EQL, buf.String()
	case "AND":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0054] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return AND, buf.String()
	case "OHR":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0055] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return OHR, buf.String()	
	case "NOT":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0056] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return NOT, buf.String()
	case "TRUE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0057] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return TRUE, buf.String()
	case "FALSE":
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0058] SCANNER: %v\n", strings.ToUpper(buf.String()))
		}
		return FALSE, buf.String()
	default:		
		if s.isDebug == "true" {
		fmt.Fprintf(s.hw, "[X0059] SCANNER: FOUND SPECIAL TOKEN: %v\n", buf.String())
		}		
		return SPECIAL_TOKEN, buf.String()
	}

	// Otherwise return as a regular identifier.
	return IDENTIFIER, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0060] SCANNER: s.read()\n")
	}
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0061] SCANNER: ch: %v\n", string(ch))
	}

	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { 
	if s.isDebug == "true" {
	fmt.Fprintf(s.hw, "[X0062] SCANNER: s.unread()\n")
	}
	_ = s.r.UnreadRune() 
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { 
	return ch == ' ' || ch == '\t' || ch == '\n' 
}

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') 
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { 
	return (ch >= '0' && ch <= '9') 
}

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
