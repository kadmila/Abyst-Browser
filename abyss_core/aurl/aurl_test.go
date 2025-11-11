package aurl

import (
	"fmt"
	"testing"
)

func ParsePrintAURL(aurl string) {
	fmt.Println(aurl)

	_aurl, err := TryParse(aurl)
	if err != nil {
		fmt.Println(err.Error() + "\n")
		return
	}
	fmt.Println("Hash: " + _aurl.Hash)
	fmt.Print("Addr: ")
	for i, c := range _aurl.Addresses {
		fmt.Print("[", i, "] "+c.String()+" ")
	}
	fmt.Println("")
	fmt.Println("Path:" + _aurl.Path)
	fmt.Println(_aurl.ToString() + "\n")
}
func TestAurl(t *testing.T) {
	ParsePrintAURL("abyss:abc:9.8.7.6:1605/somepath")
	ParsePrintAURL("abyss:abc:[2001:db8:85a3:8d3:1319:8a2e:370:7348]:443|9.8.7.6:1605/somepath")
	ParsePrintAURL("abyss:abc/somepath")
	ParsePrintAURL("abyss:abc:9.8.7.6:1605")
	ParsePrintAURL("abyss:abc")
	ParsePrintAURL("abyss:hhh:1.2.3.4:1605/")
	ParsePrintAURL("abyss:hhh:127.0.0.1:100/")

	ParsePrintAURL("abyss:hhh:www.google.com:100/")
	ParsePrintAURL("http:abc")
	ParsePrintAURL("https://abc")
	ParsePrintAURL("www.google.com")
	ParsePrintAURL("abyss://www.google.com")
	ParsePrintAURL("abyss:www.google.com")
	ParsePrintAURL("abyss:hhh:1.2.3.4.5:90/")
	ParsePrintAURL("abyss:hhh:|/")
	ParsePrintAURL("abyss:hhh::1605/")
}
