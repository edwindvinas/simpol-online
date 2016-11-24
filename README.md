# Simpol

Simpol is a Golang interpreter for the sample language called Simpol. It was patterned from github.com/edwindvinas/simpol.

This online interpreter allows users to test Simpol language online. It uses Google Appengine, Golang, and hand-written scanning and parsing logic. Please note that this is not the ideal solution for interpreters which usually use tools such as EBNF, Lex, Yacc, etc in order to create a robust interpreter software. But using these tools are complex to implement in a university project.

## Demo
http://simpol-online.appspot.com/p/c300ab584e031bcbed62eaec49b0f19f36523df6

## Installation
Requires Go.
```
$ git clone https://github.com/edwindvinas/simpol-online.git
$ //Create an Appengine project ID
$ //Update the app.yaml to indicate your project id
$ ./updategae.sh
```

## Tools
```
$ //Update the debug message markers
$ ./formatcodes.sh
$ //Update appengine project
$ ./updategae.sh
$ //Update github
$ ./updategithub.sh
```


## Examples

```bash
variable {
STG str
STG name
INT num1
INT num2
INT num3
BLN bol1
BLN bol2
}

code {
PUT $The result is: $ IN str
ASK name
PUT true IN bol1
PUT false IN bol2
PUT ADD 1 2 IN num1
PUT 100 IN num2

PRT $Your name is $
PRT name
PRT OHR true AND bol1 bol2
PUT MUL 10 ADD num1 num2 IN num3
PRT num3
PRT DIV MUL 10 ADD num1 num2 MUL 10 ADD num1 num2
PRT $Goodbye!$
}
```

## Limitations
Note that as of this initial version, the interpreter cannot handle nested conditions yet. The given code above was converted first to a un-nested format.

```bash
/* Un-nested SIMPOL Program */

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
}
```

## Sample Result

```bash
///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////
// Simpol Interpreter v1                                 //
///////////////////////////////////////////////////////////
// Author: Edwin D. Vinas                                //
// URL: http://simpol-online.appspot.com/                //
// Git: https://github.com/edwindvinas/simpol-online     //
// Debug: OFF                                            //
///////////////////////////////////////////////////////////
*********************** START RESULTS *********************
Your name is 
Edwin D. Vinas
true
1030
1
Goodbye!
*********************** END OF RESULTS ********************
///////////////////////////////////////////////////////////
// THANK YOU FOR USING SIMPOL!                           //
///////////////////////////////////////////////////////////
```

To try this online, see http://simpol-online.appspot.com/
See the explore section to see sample codes and execute them

To see the Simpol Language Specifications, please see
https://github.com/edwindvinas/simpol/blob/master/SIMPOL_SPECS.md

