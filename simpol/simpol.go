/////////////////////////////////////////////////////////////////////////////////////////////////
// SIMPOL INTERPRETER
// A university project to create an interpreter for a sample SIMPOL language.
// COPYRIGHT (c) 2016 Edwin D. Vinas
/////////////////////////////////////////////////////////////////////////////////////////////////
//REV ID: 		D0001
//REV DATE: 	2016-Nov-20
//REV DESC:		Created initial interpreter using Golang
//REV AUTH:		Edwin D. Vinas
/////////////////////////////////////////////////////////////////////////////////////////////////
//REV ID: 		D0001
//REV DATE: 	2016-Nov-24
//REV DESC:		Created initial version (w/o nested conditions)
//REV AUTH:		Edwin D. Vinas
/////////////////////////////////////////////////////////////////////////////////////////////////
package simpol

//Import packages
import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"crypto/sha1"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"text/scanner"
	"bytes"
)

//Storage of sample shared codes
type Record struct {
	Code string
}

//HTML Template
var t = template.Must(template.ParseFiles("tmpl/index.tpl"))

//HTTP handlers
func init() {
	http.HandleFunc("/explore", serveExplore)
	http.HandleFunc("/api/play", serveApiPlay)
	http.HandleFunc("/api/save", serveApiSave)
	http.HandleFunc("/p/", servePermalink)
	http.HandleFunc("/", servePermalink)
}

//Initialize
func initializeSimpol() {
	FL_CURR_VARS_SECTION = false
	FL_CURR_CODE_SECTION = false
	FL_OPEN_CURLY_BRACKET = false
	FL_CLOSE_CURLY_BRACKET = false
	FL_SAW_VARIABLE_DECL = false
	THIS_CURR_VAR_IDENT = ""
	THIS_CURR_VAR_CTR = 0	
}

//Save codes
func serveApiSave(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("serveApiSave()")
	code := r.FormValue("code")
	h := sha1.New()
	fmt.Fprintf(h, "%s", code)
	hid := fmt.Sprintf("%x", h.Sum(nil))
	key := datastore.NewKey(c, "Simpol", hid, 0, nil)
	_, err := datastore.Put(c, key, &Record{code})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", key.StringID())
}

func serveApiPlay(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("serveApiPlay()")
	
	code := r.FormValue("code")
	input := fmt.Sprintf("%v", r.FormValue("input"))
	isDebug := r.FormValue("debug")
	c.Infof("debug: %v", isDebug)
	
	initializeSimpol();
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	fmt.Fprintf(w, "// Simpol Interpreter v1                                 //\n")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	fmt.Fprintf(w, "// Author: Edwin D. Vinas                                //\n")
	fmt.Fprintf(w, "// URL: http://simpol-online.appspot.com/                //\n")
	fmt.Fprintf(w, "// Git: https://github.com/edwindvinas/simpol-online     //\n")
	if isDebug == "true" {
		//catch all debug info
		fmt.Fprintf(w, "// Debug: ON                                             //\n")
	} else {
		fmt.Fprintf(w, "// Debug: OFF                                            //\n")
	}
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	fmt.Fprintf(w, "*********************** START RESULTS *********************\n")
	if isDebug == "true" {
		fmt.Fprintf(w, "[S0001]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0002] Your input is:\n")
		fmt.Fprintf(w, "[S0003]----------------------------------------------------------\n")
		fmt.Fprintf(w, "%v\n", code)
		fmt.Fprintf(w, "[S0004]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0005] Your parameters:\n")
		fmt.Fprintf(w, "[S0006]----------------------------------------------------------\n")
		fmt.Fprintf(w, "%v\n", input)
	}
	
	//remove comment lines
	if isDebug == "true" {
	fmt.Fprintf(w, "[S0011] SIMPOL: Scanning initial codes...\n")
	}
	_, code_read, err := readInputCodes(w,r,isDebug,code)
	if isDebug == "true" {
		fmt.Fprintf(w, "[S0001]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0002] Scanned input codes:\n")
		fmt.Fprintf(w, "[S0003]----------------------------------------------------------\n")
		fmt.Fprintf(w, "%v\n", code_read)
	}
	
	//remove comment lines
	if isDebug == "true" {
	fmt.Fprintf(w, "[S0011] SIMPOL: Cleaning comments from code...\n")
	}
	ccode, err := cleanInputCodes(w,r,isDebug,code)
	if isDebug == "true" {
		fmt.Fprintf(w, "[S0001]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0002] Cleaned input codes:\n")
		fmt.Fprintf(w, "[S0003]----------------------------------------------------------\n")
		fmt.Fprintf(w, "%v\n", ccode)
	}
	if err != nil {
	fmt.Fprintf(w, "[S0011] SIMPOL: %v\n", err)
	}
	
	//store each input to variables table
	SPL := strings.Split(input, "|")
	c.Infof("input: %v", input)
	c.Infof("SPL: %v", SPL)
	for i := 1; i < len(SPL) && len(SPL) > 0; i++ {
		//this is a valid variable indentifier
		c.Infof("SPL[%v]: %v", i, SPL[i])
		
		if isDebug == "true" {
		fmt.Fprintf(w, "[S0011] SIMPOL: stored ask parameter: %v\n", SPL[i])
		}
		SPM := strings.Split(SPL[i], "=")
		THIS_CURR_VAR_CTR++
		v := new(VariableBlockTable)
		v.VarRefNum = THIS_CURR_VAR_CTR
		v.VarKeyword = THIS_CURR_VAR_IDENT
		v.VarName = SPM[0]
		v.VarValue = randSeq(16)
		//keyID
		cKey := fmt.Sprintf("%v", v.VarValue)
		putBytesToMemcacheWithExp(w,r,isDebug,cKey,[]byte(SPM[1]),3600)
		//variable name store map to key
		cKey2 := fmt.Sprintf("%v", v.VarName)
		putBytesToMemcacheWithExp(w,r,isDebug,cKey2,[]byte(cKey),3600)		
	}
	
	if isDebug == "true" {
		fmt.Fprintf(w, "[S0031]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0032] Starting parser process:\n")
		fmt.Fprintf(w, "[S0033]----------------------------------------------------------\n")
	}
	vbr, cbr, err := NewParser(w,r,isDebug,strings.NewReader(ccode)).Parse()
	if err != nil {
		fmt.Fprintf(w, "[S0034]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0035] ERROR: %v\n", err)
		fmt.Fprintf(w, "[S0036]----------------------------------------------------------\n")
		return
	}
	if isDebug == "true" {
		fmt.Fprintf(w, "[S0037]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0038] VBR: %v\n", vbr)
		fmt.Fprintf(w, "[S0039] CBR: %v\n", cbr)
		fmt.Fprintf(w, "[S0040]----------------------------------------------------------\n")
		//display vbr table
		fmt.Fprintf(w, "[S0041]----------------------------------------------------------\n")
		fmt.Fprintf(w, "[S0042] VARIABLES TABLE\n")
		fmt.Fprintf(w, "[S0043]----------------------------------------------------------\n")
		for i, v := range vbr {
			fmt.Fprintf(w, "VBR[%v]: ROW: %v\n", i, v)
		}
	}
	
	fmt.Fprintf(w, "*********************** END OF RESULTS ********************\n")

	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	fmt.Fprintf(w, "// THANK YOU FOR USING SIMPOL!                           //\n")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////\n")
	memcache.Flush(c)
	
}

//List shared codes
func serveExplore(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("serveExplore()")
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<font face=\"Courier New\" size=\"3\">")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////<br>")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////<br>")
	fmt.Fprintf(w, "// Simpol Interpreter v1                                 <br>")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////<br>")
	fmt.Fprintf(w, "// Author: Edwin D. Vinas                                <br>")
	fmt.Fprintf(w, "// URL: <a href=\"http://simpol-online.appspot.com/\">http://simpol-online.appspot.com/</a>                <br>")
	fmt.Fprintf(w, "// Git:  <a href=\"https://github.com/edwindvinas/simpol-online\">https://github.com/edwindvinas/simpol-online</a>     <br>")
	fmt.Fprintf(w, "///////////////////////////////////////////////////////////<br>")
	fmt.Fprintf(w, "***********************  CODE SAMPLES *********************<br>")
	
	//Select all shared codes
	q := datastore.NewQuery("Simpol").KeysOnly()
	recCount,_ := q.Count(c)
	if recCount > 0 {
		keys, err := q.GetAll(c, nil)
		if err != nil {
			fmt.Fprintf(w, "Error retrieving samples<br>")
			return
		}
		c := 0
		
		for _, key := range keys{
			c++
			SPL := strings.Split(fmt.Sprintf("%v", key),",")
			fmt.Fprintf(w, "[%v] <a href=\"/p/%v\">%v</a><br>", c, fmt.Sprintf("%v", SPL[1]), fmt.Sprintf("%v", SPL[1]))
		}
		fmt.Fprintf(w, "<br>%v results found!<br>", c)
	} else {
		fmt.Fprintf(w, "0 results found!<br>")
	}
	fmt.Fprintf(w, "</font>")

}

//Serve permanent link
func servePermalink(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("servePermalink()")
	path := r.URL.Path
	var code string
	if len(path) > 3 {
		id := path[3:]
		c := appengine.NewContext(r)
		var record Record
		err := datastore.Get(c, datastore.NewKey(c, "Simpol", id, 0, nil), &record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		code = record.Code
	} else {
		
code = `/* Un-nested SIMPOL Program */

/* Variables Section */
variable {
	STG str
	STG name
	INT num1
	INT num2
	INT num3
	INT num4
	INT num5
	INT num6
	INT mul1
	INT mul2
	INT div1
	BLN bol1
	BLN bol2
	BLN bol3
	BLN ohr1
}

/* Codes Section */
code {
	PUT $The result is: $ IN str
	ASK name
	
	PUT true 	IN bol1
	PUT false 	IN bol2
	ADD 1 2 	IN num1
	PUT 100 	IN num2
	
	PRT $Your name is $
	PRT name
	/*PRT OHR true AND bol1 bol2*/
	AND bol1 bol2 IN bol3
	OHR true bol3 IN ohr1
	PRT ohr1
	
	/*PUT MUL 10 ADD num1 num2 IN num3*/
	/*PUT MUL 10 (ADD num1 num2) IN num3*/
	/*                 num4             */
	ADD num1 num2 	IN num4
	MUL 10 num4 	IN num3
	PRT num3
	/*PRT DIV MUL 10 ADD num1 num2 MUL 10 ADD num1 num2*/
	/*PRT DIV (MUL 10 (ADD num1 num2)) (MUL 10 (ADD num1 num2))*/
	/*                      num6                     num5      */
	/*                 mul1                     mul2           */
	ADD num1 num2 	IN num5
	ADD num1 num2 	IN num6
	MUL 10 num5 	IN mul2
	MUL 10 num6 	IN mul1
	DIV mul1 mul2 	IN div1
	PRT div1
	
	PRT $Goodbye!$
}`
	}

	err := t.Execute(w, &struct{ Code string }{Code: code})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
					
// Function cleanInputCodes() reads the input codes given by user
func cleanInputCodes(w http.ResponseWriter, r *http.Request, isDebug string, code string) (res string, err error) {
	if isDebug == "true" {
	fmt.Fprintf(w, "[S0091]----------------------------------------------------------\n")
	fmt.Fprintf(w, "[S0092] Running function cleanInputCodes()........\n")
	fmt.Fprintf(w, "[S0093]----------------------------------------------------------\n")
	}

	//find user to host mapping		
	var buf bytes.Buffer	
	temp := strings.Split(code,"\n")
	if len(temp) > 0 {
		for j := 0; j < len(temp); j++ {
			//remove comments
			a := strings.Index(temp[j], "/*")
			b := strings.Index(temp[j], "*/")
			if a != -1 && b != -1 {
				//donothing
			} else {
				buf.WriteString(fmt.Sprintf("%v\n", temp[j]))
			}
		}
	} else {
		return buf.String(), fmt.Errorf("ERROR: No codes processed!")
	}
	
	return buf.String(), nil
}

// Function readInputCodes() reads the input codes given by user
func readInputCodes(w http.ResponseWriter, r *http.Request, isDebug string, code string) (LXR []*LexerTable, res string, err error) {
	if isDebug == "true" {
	fmt.Fprintf(w, "[S0091]----------------------------------------------------------\n")
	fmt.Fprintf(w, "[S0092] Running function readInputCodes()........\n")
	fmt.Fprintf(w, "[S0093]----------------------------------------------------------\n")
	}

	p := new(LexerTable)

 	var s scanner.Scanner
	s.Init(strings.NewReader(code))
	var tok rune
	var buf bytes.Buffer
	for tok != scanner.EOF {
		tok = s.Scan()
		if isDebug == "true" {
			//fmt.Fprintf(w, "[S0094] LINE %v: %v\n", s.Pos(), s.TokenText())
			buf.WriteString(fmt.Sprintf("LINE %v: %v\n", s.Pos(), s.TokenText()))
		}
		//store in lexer table
		p.LinePosition = fmt.Sprintf("%v", s.Pos())
		p.LexerLexeme = fmt.Sprintf("%v", s.TokenText())
		LXR = append(LXR, p)
	}
	
	return LXR, buf.String(), nil
}

