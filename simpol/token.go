//List all valid tokens or keywords and special symbols
//The parser/scanner encounters such symbols and returns value
//It is very important in parsing and interpreting the program
package simpol

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENTIFIER // main
	STRING_LITERAL //between dollars

	// Misc characters
	ASTERISK // *
	COMMA    // ,

	// Keywords in variable section
	//variable
	VARIABLE
	STG
	INT
	BLN
	
	//Keywords in codes section
	//code
	CODE
	PUT
	ASK
	PRT
	
	//Arithmetic operations
	ADD
	SUB
	MUL
	DIV
	MOD
	
	//Numeric Predicates
	GRT
	GRE
	LET
	LEE
	EQL
	
	//Logic operators
	AND
	OHR
	NOT
	
	//Boolean keywords
	TRUE
	FALSE
	
	//Predicates
	IN
	
 	//Special constants
	OPEN_CURLY_BRACKET
	CLOSE_CURLY_BRACKET
	SPECIAL_TOKEN
	DOLLAR_SYMBOL
	
	//Comments
	SLASH_SYMBOL
)
