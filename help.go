package main



//strings for displaying help to users
var helpFile = `
GoPrintLabels 

Receive message from plc or any tcp client to  server 
containing serialNumber, weight, printer ip address, 
model rev level and final / interim flag.

Usage:
	GoPrintLabels [arguments]
	No arguments are used for normal operation

Arguments:
	plcMapping	Show plc mapping for message
	string		Show comma delimited string format
	log		Log progress to console
	final		Set final label print code location
			ex. final path\to\folder
	interim		Set interim label print code location
			ex. interim path\to\folder
	config		Shows current configuration
		

`

var plcMapping = `

PLC mapping for INT or DINT array:
	Element  0 - Serial number character 1.
	Element  1 - Serial number character 2.
	Element  2 - Serial number character 3.
	Element  3 - Serial number character 4.
	Element  4 - Serial number character 5.
	Element  5 - Serial number character 6.
	Element  6 - Serial number character 7.
	Element  7 - Serial number character 8.
	Element  8 - Serial number character 9.
	Element  9 - Serial number character 10.
	Element 10 - Serial number character 11.
	Element 11 - Serial number character 12.

	Element 12 contains weight value whole number.
	Element 13 contains weight value right of the decimal.

	Element 14 - Target printer IP address 1st octet
	Element 15 - Target printer IP address 2ne octet
	Element 16 - Target printer IP address 3rd octet
	Element 17 - Target printer IP address 4th octet

	Element 18 - Model number rev level.

	Element 19 - final/interim flag.
    	A zero in this element will use final label print code by alpha code.
    	A non zer0 in this element will use interim label print code by identifier number.
    	Example - value of 1 will use print code 1.txt from interim print code location.


`

var stringMapping = `
Comma delimited string can be sent from plc message
	or tcp client using the following format:

	serial number, weight, printer IP address, rev level, interim flag
	ex - snxxxxxxxxx,wxx.xx,xxx.xxx.xxx.xxx,xx,x


`
