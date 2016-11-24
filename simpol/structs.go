package simpol

//use for gen random string
var lettersNumbers = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

//variable block parsing flags
var FL_CURR_VARS_SECTION bool
var FL_CURR_CODE_SECTION bool
var FL_OPEN_CURLY_BRACKET bool
var FL_CLOSE_CURLY_BRACKET bool
var FL_SAW_VARIABLE_DECL bool
var THIS_CURR_VAR_IDENT string
var THIS_CURR_VAR_CTR int

//code block parsing flags
var FL_SAW_CODE_DECL bool

// Lexeme Table
type LexerTable struct {
	LinePosition string 	`json:"linePosition,omitempty"`
	LexerLexeme string `json:"lexerLexeme,omitempty"`
}

// Variable Section Table
type VariableBlockTable struct {
	VarRefNum int `json:"varRefNum,omitempty"`
	VarKeyword string `json:"varKeyword,omitempty"`
	VarName string `json:"varType,omitempty"`
	VarValue string `json:"VarStrVal,omitempty"`
}

// Code Section Table
type CodeBlockTable struct {
	CodeRefNum int `json:"codeRefNum,omitempty"`
	CodeKeyword string `json:"codeKeyword,omitempty"` //PUT, PRT, ASK
	SyntaxStatus bool `json:"procStatus,omitempty"` //syntax check
	InterStatus bool `json:"interStatus,omitempty"` //interpreter status
	CodeExpr string `json:"codeExpr,omitempty"` //actual contents
	
}

// Debug Table
type OperationsTable struct {
	OpsRefNum int `json:"opsRefNum,omitempty"`
	ExecStatus bool `json:"execStatus,omitempty"`
	OpsExpr int `json:"opsExpr,omitempty"`
	DebugMessage int `json:"debugMessage,omitempty"`
	
}
