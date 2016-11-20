# Simpol Online

Simpol online is an online Golang interpreter for the sample language called Simpol. It was related to github.com/edwindvinas/simpol.

## Installation
Requires Google Appengine. See https://cloud.google.com/appengine/docs/go/

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

To see the Simpol Language Specifications, please see https://github.com/edwindvinas/simpol/blob/master/SIMPOL_SPECS.md

