package iso8583

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type encodeBase int

type iso8583FieldType int

type codepageType int

const (
	ascii encodeBase = iota
	bcd
	rbcd
)

const (
	numeric iso8583FieldType = iota
	alpha
	binary
	llvar
	llnumberic
	lllvar
	lllnumberic
)

const (
	defaultCp codepageType = iota
	windows874
	tis620
)

type tag struct {
	isMti          bool
	isSecondBitmap bool
	field          int
	length         int
	lenEncode      encodeBase
	encode         encodeBase
	fieldType      iso8583FieldType
	codePage       codepageType
}

type innerTag struct {
	field    int
	length   int
	codePage codepageType
}

func parseTag(f reflect.StructField) (t tag, err error) {
	if strings.ToLower(f.Name) == "mti" {
		t.isMti = true
		t.encode = parseEncode(f.Tag.Get("encode"))
		return
	}
	if strings.ToLower(f.Name) == "secondbitmap" {
		t.isSecondBitmap = true
		return
	}
	if t.field, err = strconv.Atoi(f.Tag.Get("field")); err != nil {
		err = fmt.Errorf("field number must be specified")
		return
	}
	if t.length, err = strconv.Atoi(f.Tag.Get("length")); err != nil {
		t.length = -1
	}
	if raw := f.Tag.Get("encode"); raw != "" {
		enc := strings.Split(raw, ",")
		if len(enc) == 2 {
			t.lenEncode = parseEncode(enc[0])
			t.encode = parseEncode(enc[1])
		} else {
			t.encode = parseEncode(enc[0])
		}
	} else {
		t.encode = parseEncode(raw)
	}
	if t.fieldType, err = parseType(f.Tag.Get("type")); err != nil {
		return
	}
	t.codePage, err = parseCodepage(f.Tag.Get("cp"))
	return
}

func parseInnerTag(rt reflect.StructTag) (t innerTag, err error) {
	if t.field, err = strconv.Atoi(rt.Get("field")); err != nil {
		return
	}
	if t.length, err = strconv.Atoi(rt.Get("length")); err != nil {
		return
	}
	t.codePage, err = parseCodepage(rt.Get("cp"))
	return
}

func parseEncode(s string) encodeBase {
	switch strings.ToLower(s) {
	case "ascii":
		return ascii
	case "lbcd":
		fallthrough
	case "bcd":
		return bcd
	case "rbcd":
		return rbcd
	}
	return ascii
}

func parseType(s string) (iso8583FieldType, error) {
	switch strings.ToLower(s) {
	case "numeric":
		return numeric, nil
	case "alpha":
		return alpha, nil
	case "binary":
		return binary, nil
	case "llvar":
		return llvar, nil
	case "llnumberic":
		return lllnumberic, nil
	case "lllvar":
		return lllvar, nil
	case "lllnumberic":
		return lllnumberic, nil
	}
	return -1, fmt.Errorf("Unsupport type for type %s", s)
}

func parseCodepage(s string) (codepageType, error) {
	switch strings.ToLower(s) {
	case "":
		return defaultCp, nil
	case "tis-620":
		return tis620, nil
	case "windows-874":
		return windows874, nil
	}
	return -1, fmt.Errorf("Unsupport codepage %s", s)
}

func (c codepageType) value() string {
	switch c {
	case tis620:
		return "TIS-620"
	case windows874:
		return "Windows-874"
	default:
		return ""
	}
}
