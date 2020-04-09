package iso8583v2

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

func decodeUtf8(decodeTable [][]byte, ebc []byte) []byte {
	var uni []byte
	for _, v := range ebc {
		subArr := decodeTable[v]
		for _, b := range subArr {
			uni = append(uni, b)
		}
	}
	return uni
}

func encodeUtf8(encodeTable map[string]byte, utf []byte) []byte {
	var ebc []byte
	for _, ch := range string(utf) {
		ebc = append(ebc, encodeTable[fmt.Sprintf("%c", ch)])
	}
	return ebc
}

//DecodeUTF8 is converting byte with codepage to byte in utf8
func decodeUTF8(codePage string, b []byte) []byte {
	return defaultDecodeUtf8(codePage, b)
}

//EncodeUTF8 is converting byte in utf8 to byte in codepage
func encodeUTF8(codePage string, b []byte) []byte {
	return defaultEncodeUtf8(codePage, b)
}

func defaultDecodeUtf8(codePage string, b []byte) []byte {
	newCodeReader := bytes.NewBuffer(b)
	reader, err := charset.NewReaderLabel(codePage, newCodeReader)
	if err != nil {
		return b
	}

	nb, err := ioutil.ReadAll(reader)
	if err != nil {
		return b
	}
	return nb
}

func defaultEncodeUtf8(codePage string, b []byte) []byte {
	e, _ := charset.Lookup(codePage)
	reader := transform.NewReader(bytes.NewReader(b), e.NewEncoder())
	nb, err := ioutil.ReadAll(reader)
	if err != nil {
		return b
	}
	return nb
}
