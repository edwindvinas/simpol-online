//Parser - parses the tokens and performs hard-coded
//operations to interpret the input codes
package simpol

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"math/rand"
	"time"
	"strings"
	"appengine/memcache"
	"appengine"
	"bytes"
	"strconv"
)

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	hw  http.ResponseWriter
	hr  *http.Request
	isDebug string
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(w http.ResponseWriter, r *http.Request, isDebug string, b io.Reader) *Parser {
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0001] NewParser()\n")
	}
	return &Parser{s: NewScanner(w,r,isDebug,b), hw: w, hr: r, isDebug: isDebug}
}

// Parse parses the SIMPOL statement
func (p *Parser) Parse() (vbr []*VariableBlockTable, cbr []*CodeBlockTable, err error) {
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0002] PARSER: Parse()\n")
	}

	//Remove any comments
	if tok, lit := p.scanIgnoreComments(); lit != "/" {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0003] PARSER: ERROR: found lit: %v tok: %v, expected / \n", lit, tok)
		}
		//repeat
		if tok, lit := p.scanIgnoreComments(); lit != "/" {
			if p.isDebug == "true" {
			fmt.Fprintf(p.hw, "[P0004] PARSER: ERROR: found lit: %v tok: %v, expected / \n", lit, tok)
			}
			return nil, nil, fmt.Errorf("found %q, expected /", lit)
		}
	}
	
	// First token should be a "VARIABLE" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != VARIABLE {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0005] PARSER: ERROR: found lit: %v tok: %v, expected VARIABLE declaration\n", lit, tok)
		}
		if tok, lit := p.scanIgnoreWhitespace(); tok != VARIABLE {
			fmt.Fprintf(p.hw, "[P0006] PARSER: ERROR: found lit: %v tok: %v, expected VARIABLE declaration\n", lit, tok)
			return nil, nil, fmt.Errorf("found %q, expected VARIABLE declaration", lit)
		}
	}
	
	// Next we should expect a "{"
	if tok, lit := p.scanIgnoreWhitespace(); tok != OPEN_CURLY_BRACKET {
		fmt.Fprintf(p.hw, "[P0009] PARSER: ERROR: found lit: %v tok: %v, expected OPEN_CURLY_BRACKET \n", lit, tok)
		return nil, nil, fmt.Errorf("found %q, expected OPEN_CURLY_BRACKET", lit)
	} else {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0010] !!!!!!!!! START VARIABLE BLOCK SECTION !!!!!!!!!!!\n")
		}
		FL_OPEN_CURLY_BRACKET = true
	}
	
	// Next we should loop over all our variables in the VB block
	for {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0011] PARSER: loop\n")
		fmt.Fprintf(p.hw, "[P0012] PARSER: p.scanIgnoreWhitespace()\n")
		}
		tok, lit := p.scanIgnoreWhitespace()
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0013] PARSER: tok: %v lit: %v\n", tok, lit)
		}

		if tok == CLOSE_CURLY_BRACKET {
			FL_CLOSE_CURLY_BRACKET = false
			if p.isDebug == "true" {
			fmt.Fprintf(p.hw, "[P0015] PARSER: !!!!!!!!! CLOSE VARIABLE SECTION !!!!!!!!!\n")
			}
			break
		}
		
		if tok == IDENTIFIER && FL_SAW_VARIABLE_DECL == true{
			if p.isDebug == "true" {
			fmt.Fprintf(p.hw, "[P0016] PARSER: found valid variable identifier: %v\n", lit)
			}
			//check correct naming convention
			//should not start with number 0-9, ., or special char
			match, _ := regexp.MatchString("[a-zA-Z]?+", lit)
			if match == false {
			fmt.Fprintf(p.hw, "[P0017] PARSER: ERROR: variable should not start with number or special char; lit: %v\n", lit)
			return nil, nil, fmt.Errorf("variable should not start with number or special char; lit: %v\n", lit)
			}
			
			//if this an ASK variable; skip initialization
			isFound := getStringValue(p.hw, p.hr,p.isDebug,lit)
			if isFound == "" {
				//store this identifier
				FL_SAW_VARIABLE_DECL = false
				THIS_CURR_VAR_CTR++
				v := new(VariableBlockTable)
				v.VarRefNum = THIS_CURR_VAR_CTR
				v.VarKeyword = THIS_CURR_VAR_IDENT
				v.VarName = lit
				v.VarValue = randSeq(16)
				//value store
				cKey := fmt.Sprintf("%v", v.VarValue)
				thisValue := "INITIALIZED"
				if THIS_CURR_VAR_IDENT == "INT" {
					thisValue = "0"
				} else if THIS_CURR_VAR_IDENT == "BLN" {
					thisValue = "false"
				}
				putBytesToMemcacheWithExp(p.hw,p.hr,p.isDebug,cKey,[]byte(thisValue),3600)
				//variable name store
				cKey2 := fmt.Sprintf("%v", v.VarName)
				putBytesToMemcacheWithExp(p.hw,p.hr,p.isDebug,cKey2,[]byte(cKey),3600)
				vbr = append(vbr, v)
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0018] PARSER: Saved storage: varName: %v varKey: %v\n", v.VarName, v.VarValue)
				fmt.Fprintf(p.hw, "[P0019] PARSER: THIS_CURR_VAR_IDENT: %v\n", THIS_CURR_VAR_IDENT)
				}	
			}
		}
		
		if tok, lit := p.scanIgnoreWhitespace(); tok == STG || tok == INT || tok == BLN  {
			fmt.Fprintf(p.hw, "[P0021] PARSER: ERROR: found lit: %v tok: %v, unexpected STG, INT, or BLN \n", lit, tok)
			return nil, nil, fmt.Errorf("found %q, unexpected STG, INT, or BLN", lit)
		} else {
			FL_SAW_VARIABLE_DECL = false
			//this is a valid variable indentifier
			if p.isDebug == "true" {
			fmt.Fprintf(p.hw, "[P0022] PARSER: found valid variable identifier: %v\n", lit)
			}
			//if this an ASK variable; skip initialization
			isFound := getStringValue(p.hw, p.hr,p.isDebug,lit)
			if isFound == "" {
				//store in vbr
				//store this identifier
				FL_SAW_VARIABLE_DECL = false
				//save to VBR table
				THIS_CURR_VAR_CTR++
				v := new(VariableBlockTable)
				v.VarRefNum = THIS_CURR_VAR_CTR
				v.VarKeyword = THIS_CURR_VAR_IDENT
				v.VarName = lit
				v.VarValue = randSeq(16)
				cKey := fmt.Sprintf("%v", v.VarValue)
				thisValue := "INITIALIZED"
				if THIS_CURR_VAR_IDENT == "INT" {
					thisValue = "0"
				} else if THIS_CURR_VAR_IDENT == "BLN" {
					thisValue = "false"
				}
				putBytesToMemcacheWithExp(p.hw,p.hr,p.isDebug,cKey,[]byte(thisValue),3600)
				//variable name store
				cKey2 := fmt.Sprintf("%v", v.VarName)
				putBytesToMemcacheWithExp(p.hw,p.hr,p.isDebug,cKey2,[]byte(cKey),3600)		
				vbr = append(vbr, v)
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0023] PARSER: Saved storage: varName: %v varKey: %v\n", v.VarName, v.VarValue)
				fmt.Fprintf(p.hw, "[P0024] PARSER: THIS_CURR_VAR_IDENT: %v\n", THIS_CURR_VAR_IDENT)
				}
			}
		}
	}
	
	//Now, lets can the code section
	//Expect here to find the CODE declaration section
	if tok, lit := p.scanIgnoreWhitespace(); tok != CODE {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0026] PARSER: ERROR: found lit: %v tok: %v, expected CODE declaration\n", lit, tok)
		}
		if tok, lit := p.scanIgnoreWhitespace(); tok != CODE {
			fmt.Fprintf(p.hw, "[P0027] PARSER: ERROR: found lit: %v tok: %v, expected CODE declaration\n", lit, tok)
			return nil, nil, fmt.Errorf("found %q, expected CODE declaration", lit)
		}
	}
	
	//Next, lets expect a "{" to open code section
	if tok, lit := p.scanIgnoreWhitespace(); tok != OPEN_CURLY_BRACKET {
		fmt.Fprintf(p.hw, "[P0028] PARSER: ERROR: found lit: %v tok: %v, expected OPEN_CURLY_BRACKET \n", lit, tok)
		//}
		return nil, nil, fmt.Errorf("found %q, expected OPEN_CURLY_BRACKET", lit)
	} else {
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0029] !!!!!!!!! START CODE BLOCK SECTION !!!!!!!!!!!\n")
		}
		FL_OPEN_CURLY_BRACKET = true
	}	
	
	// Next we should loop over all our codes in the CODES section
	for {
		
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0030] PARSER: loop\n")
		fmt.Fprintf(p.hw, "[P0031] PARSER: p.scanIgnoreWhitespace()\n")
		}
		tok, lit := p.scanIgnoreWhitespace()
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0032] PARSER: tok: %v lit: %v\n", tok, lit)
		}
		
		switch tok {
			
			case ADD:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0033] PARSER: /////// ADD HANDLER ////////\n")
				}
				//expects 
				//ADD val1 IN val2
				//ADD val1 val2 val3 IN val4
				tempAcc := 0
				for {
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0034] PARSER: tok: %v lit: %v\n", tok, lit)
					}
					tempNum, err := strconv.Atoi(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0035] PARSER: ADD tempNum: %v\n", tempNum)
					}
					if err != nil && tok != IN {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0036] PARSER: ADD OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					}
					tempAcc = tempAcc + tempNum;
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0037] PARSER: ADD TEMPACC: %v\n", tempAcc)
					}
					if tok == IN {
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0038] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0039] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0040] PARSER: ADD OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						tempAcc = tempAcc + tempNum
						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", tempAcc))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0041] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, tempAcc)
						}						
						break
					}
				}
			
			case DIV:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0042] PARSER: /////// DIV HANDLER ////////\n")
				}
				//expects 
				//DIV val1 val2 IN divTotal
				//DIV 100 val1 IN val5
				FL_IN_FOUND := false
				divOper1 := 0
				divOper2 := 0
				var divRes float64
				for {
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0043] PARSER: tok: %v lit: %v\n", tok, lit)
					}

					tempNum, err := strconv.Atoi(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0044] PARSER: DIV tempNum: %v\n", tempNum)
					}
					if err != nil && tok != IN && divOper1 == 0 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						divOper1 = tempNum
						if err != nil {
						fmt.Fprintf(p.hw, "[P0045] PARSER: DIV OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					} else if divOper2 == 0 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						divOper2 = tempNum
						if err != nil {
						fmt.Fprintf(p.hw, "[P0046] PARSER: DIV OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
					}
					if tok == IN {
						FL_IN_FOUND = true
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0047] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0048] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0049] PARSER: DIV OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0050] PARSER: divRes(%f) = divOper1(%v) / divOper2(%v)\n", divRes, divOper1, divOper2)
						}
						divRes = float64(divOper1)/float64(divOper2)
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0051] PARSER: divRes(%f) = divOper1(%v) / divOper2(%v)\n", divRes, divOper1, divOper2)
						}						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", divRes))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0052] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, divRes)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0053] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
				
			case AND:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0054] PARSER: /////// AND HANDLER ////////\n")
				}
				//expects 
				//AND val1 val2 IN andTotal
				//AND 100 val1 IN val5
				//AND true false IN val6
				FL_IN_FOUND := false
				andOper1 := false
				andOper2 := false
				andRes := false
				i := 0
				for {
					
					i++
					
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0055] PARSER: tok: %v lit: %v\n", tok, lit)
					}

					tempBool, err := strconv.ParseBool(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0056] PARSER: AND tempBool: %v\n", tempBool)
					}
					if err != nil && tok != IN && i == 1 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempBool, err := strconv.ParseBool(varValue)
						andOper1 = tempBool
						if err != nil {
						fmt.Fprintf(p.hw, "[P0057] PARSER: AND OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					} else if i == 2 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempBool, err := strconv.ParseBool(varValue)
						andOper2 = tempBool
						if err != nil {
						fmt.Fprintf(p.hw, "[P0058] PARSER: AND OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
					}
					if tok == IN {
						FL_IN_FOUND = true
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0059] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0060] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempBool, err := strconv.ParseBool(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0061] PARSER: AND OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						andRes = tempBool
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0062] PARSER: andRes(%v) = andOper1(%v) && andOper2(%v)\n", andRes, andOper1, andOper2)
						}
						andRes = andOper1 && andOper2
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0063] PARSER: andRes(%v) = andOper1(%v) && andOper2(%v)\n", andRes, andOper1, andOper2)
						}						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", andRes))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0064] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, andRes)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0065] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}

				
			case OHR:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0066] PARSER: /////// OHR HANDLER ////////\n")
				}
				//expects 
				//OHR val1 val2 IN andTotal
				//OHR 100 val1 IN val5
				//OHR true false IN val6
				FL_IN_FOUND := false
				ohrOper1 := false
				ohrOper2 := false
				ohrRes := false
				i := 0
				for {
					
					i++
					
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0067] PARSER: tok: %v lit: %v\n", tok, lit)
					}

					tempBool, _ := strconv.ParseBool(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0068] PARSER: OHR tempBool: %v\n", tempBool)
					}
					if i == 1 && tok != IN {
						if (lit == "true" || lit == "false") {
							ohrOper1 = tempBool
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0070] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper1: %v\n", tok, lit, ohrOper1)
							}
						} else {
							//get value if exists
							varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
							tempBool, err := strconv.ParseBool(varValue)
							ohrOper1 = tempBool
							if err != nil {
							fmt.Fprintf(p.hw, "[P0069] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper1: %v\n", tok, lit, ohrOper1)
							break
							}
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0070] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper1: %v\n", tok, lit, ohrOper1)
							}
						}
						
					} else if tok != IN && i == 2 {
						if (lit != "true" || lit != "false") {
							ohrOper2 = tempBool
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0070] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper2: %v\n", tok, lit, ohrOper2)
							}
						} else {
							//get value if exists
							varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
							tempBool, err := strconv.ParseBool(varValue)
							ohrOper2 = tempBool
							if err != nil {
							fmt.Fprintf(p.hw, "[P0071] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper2: %v\n", tok, lit, ohrOper2)
							break
							}
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0072] PARSER: OHR OPERATION ERROR tok: %v lit: %v ohrOper2: %v\n", tok, lit, ohrOper2)
							}
						}
					}
					if tok == IN {
						FL_IN_FOUND = true
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0073] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0074] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempBool, err := strconv.ParseBool(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0075] PARSER: OHR OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						ohrRes = tempBool
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0076] PARSER: ohrRes(%v) = ohrOper1(%v) || ohrOper2(%v)\n", ohrRes, ohrOper1, ohrOper2)
						}
						ohrRes = ohrOper1 || ohrOper2
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0077] PARSER: ohrRes(%v) = ohrOper1(%v) || ohrOper2(%v)\n", ohrRes, ohrOper1, ohrOper2)
						}						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", ohrRes))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0078] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, ohrRes)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0079] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
				
			case NOT:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0080] PARSER: /////// NOT HANDLER ////////\n")
				}
				//expects 
				//NOT val1 IN notRes
				//NOT true IN val6
				FL_IN_FOUND := false
				notOper1 := false
				notRes := false
				i := 0
				for {
					
					i++
					
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0081] PARSER: tok: %v lit: %v\n", tok, lit)
					}

					tempBool, err := strconv.ParseBool(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0082] PARSER: NOT tempBool: %v\n", tempBool)
					}
					if err != nil && tok != IN && i == 1 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempBool, err := strconv.ParseBool(varValue)
						notOper1 = tempBool
						if err != nil {
						fmt.Fprintf(p.hw, "[P0083] PARSER: NOT OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					} else if i > 2  && tok == IN {
						//get value if exists
						fmt.Fprintf(p.hw, "[P0084] PARSER: NOT SYNTAX ERROR - MULTIPLE FOUND tok: %v lit: %v\n", tok, lit)
						break
					}
					if tok == IN {
						FL_IN_FOUND = true
						//get dest token
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0085] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0086] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempBool, err := strconv.ParseBool(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0087] PARSER: NOT OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						notRes = tempBool
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0088] PARSER: notRes(%v) = !notOper1(%v)\n", notRes, notOper1)
						}
						notRes = !notOper1
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0089] PARSER: notRes(%v) = !notOper1(%v)\n", notRes, notOper1)
						}						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", notRes))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0090] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, notRes)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0091] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
				
			case MOD:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0092] PARSER: /////// MOD HANDLER ////////\n")
				}
				//expects 
				//MOD val1 val2 IN divTotal
				//MOD 100 val1 IN val5
				FL_IN_FOUND := false
				divOper1 := 0
				divOper2 := 0
				divRes := 0
				for {
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0093] PARSER: tok: %v lit: %v\n", tok, lit)
					}

					tempNum, err := strconv.Atoi(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0094] PARSER: MOD tempNum: %v\n", tempNum)
					}
					if err != nil && tok != IN && divOper1 == 0 {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						divOper1 = tempNum
						if err != nil {
						fmt.Fprintf(p.hw, "[P0095] PARSER: MOD OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					} else if divOper2 == 0 {
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						divOper2 = tempNum
						if err != nil {
						fmt.Fprintf(p.hw, "[P0096] PARSER: MOD OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
					}
					if tok == IN {
						FL_IN_FOUND = true
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0097] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0098] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0099] PARSER: MOD OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0100] PARSER: divRes(%v) = divOper1(%v) / divOper2(%v)\n", divRes, divOper1, divOper2)
						}
						divRes = divOper1%divOper2
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0101] PARSER: divRes(%v) = divOper1(%v) / divOper2(%v)\n", divRes, divOper1, divOper2)
						}						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", divRes))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0102] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, divRes)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0103] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
			
			case SUB:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0104] PARSER: /////// SUB HANDLER ////////\n")
				}
				//expects 
				//SUB val1 IN val2
				//SUB val1 val2 val3 IN val4
				//SUB 100 val1 IN val5
				tempAcc := 0
				FL_IN_FOUND := false
				for {
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0105] PARSER: tok: %v lit: %v\n", tok, lit)
					}
					//add to temp
					tempNum, err := strconv.Atoi(lit)
					if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0106] PARSER: SUB tempNum: %v\n", tempNum)
					}
					if err != nil && tok != IN {
						//get value if exists
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0107] PARSER: SUB OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
					}
					tempAcc = tempAcc + tempNum;
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0108] PARSER: SUB TEMPACC: %v\n", tempAcc)
					}
					if tok == IN {
						FL_IN_FOUND = true
						//get dest token
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0109] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0110] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0111] PARSER: SUB OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						tempAcc = tempNum - tempAcc
						
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", tempAcc))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0112] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, tempAcc)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0113] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
				
			case MUL:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0114] PARSER: /////// MUL HANDLER ////////\n")
				}
				//expects 
				//MUL val1 IN val2
				//MUL val1 val2 val3 IN val4
				//MUL 100 val1 IN val5
				tempAcc := 1
				tempNum := 0
				FL_IN_FOUND := false
				for {
					tok, lit := p.scanIgnoreWhitespace()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0115] PARSER: tok: %v lit: %v\n", tok, lit)
					}
					if tok != IN {
						tempNum, err = strconv.Atoi(lit)
						if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0116] PARSER: MUL tempNum: %v\n", tempNum)
						}
						if err != nil {
							//get value if exists
							varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
							tempNum, err = strconv.Atoi(varValue)
							if err != nil {
							fmt.Fprintf(p.hw, "[P0117] PARSER: MUL OPERATION ERROR tok: %v lit: %v\n", tok, lit)
							break
							}
							
						}
					}
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0118] PARSER: tempAcc (%v) = tempNum (%v) * tempAcc (%v)\n", tempAcc, tempNum, tempAcc)
					}
					if tok != IN {
						tempAcc = tempAcc * tempNum;
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0119] PARSER: MUL TEMPACC: %v\n", tempAcc)
						}
					}
					if tok == IN {
						FL_IN_FOUND = true
						//get dest token
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0120] PARSER: tok: %v lit: %v\n", tok, lit)
						}
						//saveto storage
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0121] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
						}
						varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
						
						tempNum, err = strconv.Atoi(varValue)
						if err != nil {
						fmt.Fprintf(p.hw, "[P0122] PARSER: MUL OPERATION ERROR tok: %v lit: %v\n", tok, lit)
						break
						}
						
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0123] PARSER: tempAcc (%v) = tempNum (%v) * tempAcc (%v)\n", tempAcc, tempNum, tempAcc)
						}
						storeStringValue(p.hw, p.hr,p.isDebug,lit,fmt.Sprintf("%v", tempAcc))
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0124] PARSER: VARIABLE VALUE SAVED var: %v value: %v\n", lit, tempAcc)
						}						
						break
					}
				}
				
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0125] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
			
			case PUT:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0126] PARSER: /////// PUT HANDLER ////////\n")
				}
				
				FL_IN_FOUND := false
				
				//handle PUT $string here$ IN identifier
				//check if dollar is found
				tok, lit := p.scanIgnoreWhitespace()
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0127] PARSER: tok: %v lit: %v\n", tok, lit)
				}
				if tok == DOLLAR_SYMBOL {
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0128] FOUND STRING LITERAL: tok: %v lit: %v\n", tok, lit)
					}
					// scan the string literal inside dollars
					tok, lit := p.scanDollarStrings()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0129] PARSER: tok: %v lit: %v\n", tok, lit)
					}
					if tok == STRING_LITERAL {
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0130] PARSER: GOT STRING_LITERAL BETWEEN DOLLARS tok: %v lit: %v\n", tok, lit)
						}	
						//store this to the variable
						currStringValue := lit
						tok, lit := p.scanIgnoreWhitespace()
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0131] PARSER: tok: %v lit: %v\n", tok, lit)
						}	
						if tok == IN {
							FL_IN_FOUND = true
							//get the target variable
							tok, lit := p.scanIgnoreWhitespace()
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0132] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
							}
							storeStringValue(p.hw, p.hr,p.isDebug,lit, currStringValue)
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0133] PARSER: VARIABLE VALUE SAVED var: %v lit: %v\n", lit, currStringValue)
							}
							
						} else {
							fmt.Fprintf(p.hw, "[P0134] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
						}
					}
				} else {
					//this can be an identifier or an expression
					//PUT 100 IN str2
					//PUT true IN bol1
					//PUT str IN str2
					//PUT ADD 1 2 IN num1
					//PUT MUL 10 ADD num1 num2 IN num3
					
					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
					//-------- NESTED OPERATIONS ARE NOT YET WORKING ------------
					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
					switch {
						case tok == ADD:
						
						case tok == MUL:
						
						case tok == DIV:
						
						case tok == MOD:
						
						case tok == SUB:
						
						default:
							//this is a value or identifier
							//PUT 100 IN str2
							//PUT true IN bol1
							//PUT str IN str2
							varValue := ""
							_, err := strconv.Atoi(lit)
							if err != nil {
								varValue = getStringValue(p.hw, p.hr,p.isDebug,lit)
							}
							currStringValue := varValue
							if p.isDebug == "true" {
							fmt.Fprintf(p.hw, "[P0135] PARSER: VARIABLE VALUE GOT var: %v lit: %v\n", lit, varValue)
							}
							if varValue != "" {
								//PUT str IN str2
								//get the target variable
								varValue := getStringValue(p.hw, p.hr,p.isDebug,lit)
								if p.isDebug == "true" {
								fmt.Fprintf(p.hw, "[P0136] PARSER: VARIABLE VALUE GOT var: %v lit: %v\n", lit, varValue)
								}
								//get in variable
								tok, lit := p.scanIgnoreWhitespace()
								if p.isDebug == "true" {
								fmt.Fprintf(p.hw, "[P0137] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
								}
								if tok == IN {
									FL_IN_FOUND = true
									//get the target variable
									tok, lit := p.scanIgnoreWhitespace()
									if p.isDebug == "true" {
									fmt.Fprintf(p.hw, "[P0138] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
									}
									storeStringValue(p.hw, p.hr,p.isDebug,lit, currStringValue)
									if p.isDebug == "true" {
									fmt.Fprintf(p.hw, "[P0139] PARSER: VARIABLE VALUE SAVED var: %v lit: %v\n", lit, currStringValue)
									}									
								} else {
									fmt.Fprintf(p.hw, "[P0140] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
								}							
							} else {
								//PUT true IN bol1
								//PUT 100 IN str2
								//a fixed value
								//get the IN variable
								currStringValue = lit
								tok, lit := p.scanIgnoreWhitespace()
								if p.isDebug == "true" {
								fmt.Fprintf(p.hw, "[P0141] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
								}
								if tok == IN {
									FL_IN_FOUND = true
									//get the target variable
									tok, lit := p.scanIgnoreWhitespace()
									if p.isDebug == "true" {
									fmt.Fprintf(p.hw, "[P0142] PARSER: TARGET STORAGE VARIABLE tok: %v lit: %v\n", tok, lit)
									}
									storeStringValue(p.hw, p.hr,p.isDebug,lit, currStringValue)
									if p.isDebug == "true" {
									fmt.Fprintf(p.hw, "[P0143] PARSER: VARIABLE VALUE SAVED var: %v lit: %v\n", lit, currStringValue)
									}									
								} else {
									fmt.Fprintf(p.hw, "[P0144] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
								}

							}
							
					}

				}
				if FL_IN_FOUND == false {
					fmt.Fprintf(p.hw, "[P0145] PARSER: SYNTAX ERROR - EXPECTS IN MARKER: %v lit: %v\n", tok, lit)
				}
			
			case PRT:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0146] PARSER: /////// PRT HANDLER ////////\n")
				}
				
				FL_DOLLAR_FOUND := false
				
				//handle PRT $string here$ IN identifier
				//check if dollar is found
				tok, lit := p.scanIgnoreWhitespace()
				litValue := lit
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0147] PARSER: tok: %v lit: %v\n", tok, lit)
				}
				if tok == DOLLAR_SYMBOL {
					FL_DOLLAR_FOUND = true
					//value is to be printed outright!
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0148] FOUND STRING LITERAL: tok: %v lit: %v\n", tok, lit)
					}
					// scan the string literal inside dollars
					tok, lit := p.scanDollarStrings()
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0149] PARSER: tok: %v lit: %v\n", tok, lit)
					}
					if tok == STRING_LITERAL {
						if p.isDebug == "true" {
						fmt.Fprintf(p.hw, "[P0150] PARSER: GOT STRING_LITERAL BETWEEN DOLLARS tok: %v lit: %v\n", tok, lit)
						}
						//display now
						fmt.Fprintf(p.hw, "%v\n", lit)
					}
				} else {
					//this is either an identifier or an expression!
					FL_DOLLAR_FOUND = true //as if 
					varValue := getStringValue(p.hw, p.hr,p.isDebug,litValue)
					if p.isDebug == "true" {
					fmt.Fprintf(p.hw, "[P0151] PARSER: VARIABLE VALUE GOT var: %v lit: %v\n", lit, varValue)
					}
					//display now
					fmt.Fprintf(p.hw, "%v\n", varValue)
				
				}
				
				if FL_DOLLAR_FOUND == false {
					fmt.Fprintf(p.hw, "[P0152] PARSER: SYNTAX ERROR - EXPECTS $ MARKER: %v lit: %v\n", tok, lit)
				}
				
			case ASK:
				if p.isDebug == "true" {
				fmt.Fprintf(p.hw, "[P0153] PARSER: /////// ASK HANDLER ////////\n")
				}		
				//donothing; input has been processed!
		}
		
		if tok == CLOSE_CURLY_BRACKET {
			FL_CLOSE_CURLY_BRACKET = false
			if p.isDebug == "true" {
			fmt.Fprintf(p.hw, "[P0154] PARSER: !!!!!!!!! CLOSE CODE SECTION !!!!!!!!!\n")
			}
			break
		}
	}

	// Return the successfully parsed statement.
	return vbr, cbr, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0160] PARSER: p.scan\n")
	}
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		if p.isDebug == "true" {
		fmt.Fprintf(p.hw, "[P0161] PARSER: p.buf.tok: %v\n", p.buf.tok)
		fmt.Fprintf(p.hw, "[P0162] PARSER: p.buf.lit: %v\n", p.buf.lit)
		}
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0163] PARSER: p.s.Scan()\n")
	}
	tok, lit = p.s.Scan()
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0164] PARSER: tok: %v\n", tok)
	fmt.Fprintf(p.hw, "[P0165] PARSER: lit: %v\n", lit)
	
	}
	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
//func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0166] PARSER: p.scanIgnoreWhitespace()\n")
	}
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}


func (p *Parser) scanIgnoreComments() (tok Token, lit string) {
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0168] PARSER: p.scanIgnoreComments()\n")
	}
	tok, lit = p.scan()
	return
}

// scanDollarStrings scans the whole string content between two dollars
func (p *Parser) scanDollarStrings() (tok Token, lit string) {
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0169] PARSER: p.scanDollarStrings()\n")
	}
	
	var buf bytes.Buffer
	
	for {

		tok, lit = p.scan()
		if tok == DOLLAR_SYMBOL {
			//end of literal
			return STRING_LITERAL, buf.String()
			break
		} else {
			//get string
			buf.WriteString(fmt.Sprintf("%v", lit))
		}	
		
	}
	
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { 
	if p.isDebug == "true" {
	fmt.Fprintf(p.hw, "[P0170] PARSER: p.unscan()\n")
	}
	p.buf.n = 1 
}

//generate randome string
func randSeq(n int) string {
    rand.Seed(time.Now().UTC().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = lettersNumbers[rand.Intn(len(lettersNumbers))]
    }
    return "MEM_"+strings.ToLower(string(b))
}

//get storage memcache
func getStrMemcacheValueByKey(w http.ResponseWriter, r *http.Request, isDebug, cKey string) (cVal string) {
	c := appengine.NewContext(r)
	cKey = strings.TrimSpace(cKey)
	if item, err := memcache.Get(c, cKey); err == memcache.ErrCacheMiss {
		//missed
	} else if err != nil {
		//c.Errorf("error getting item: %v", err)
	} else {
		cVal = fmt.Sprintf("%s", item.Value)
	}
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0171] PARSER: keyID: %v value: %v\n", cKey, cVal)
	}
	return cVal
	
}

//set storage memcache 
func putBytesToMemcacheWithExp(w http.ResponseWriter, r *http.Request,isDebug, cKey string,sBytes []byte,tExp time.Duration) {
	c := appengine.NewContext(r)
	cKey = strings.TrimSpace(cKey)
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0172] PARSER: keyID: %v value: %v\n", cKey, string(sBytes))
	}
	item := &memcache.Item{
		Key:   cKey,
		Value: sBytes,
		Expiration: time.Second * tExp,
	}
				
	if err := memcache.Add(c, item); err == memcache.ErrNotStored {
		memcache.Set(c, item)
	}
	return
	
}

//set storage memcache 
func storeStringValue(w http.ResponseWriter, r *http.Request,isDebug,varName string, varValue string) {
	varName = strings.TrimSpace(varName)
	cKey := fmt.Sprintf("%v", varName)
	keyID := getStrMemcacheValueByKey(w,r,isDebug,cKey)
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0173] PARSER: keyID: %v\n", keyID)
	}
	putBytesToMemcacheWithExp(w,r,isDebug,keyID,[]byte(varValue), 3600)
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0174] PARSER: varValue: %v\n", varValue)
	}
}


//get storage memcache 
func getStringValue(w http.ResponseWriter, r *http.Request,isDebug, varName string) (varValue string) {
	varName = strings.TrimSpace(varName)
	cKey := fmt.Sprintf("%v", varName)
	keyID := getStrMemcacheValueByKey(w,r,isDebug,cKey)
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0175] PARSER: keyID: %v\n", keyID)
	}
	varValue = getStrMemcacheValueByKey(w,r,isDebug,keyID)
	if isDebug == "true" {
	fmt.Fprintf(w, "[P0176] PARSER: keyValue: %v\n", varValue)
	}
	return
	
}
