package iso8583v2

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
	mtiWord = "mti"

	encodeWord     = "encode"
	fieldWord      = "field"
	lengthWord     = "length"
	bitmapsizeWord = "bitmapsize"
	typeWord       = "type"
	codepageWord   = "cp"

	fixedFieldWord    = "field"
	fixedLengthWord   = "length"
	fixedCodepageWord = "cp"
)

const (
	ascii encodeBase = iota + 1
	bcd
	rbcd
)

const (
	numeric iso8583FieldType = iota + 1
	alpha
	binary
	llvar
	lllvar
)

const (
	defaultCp codepageType = iota + 1
	windows874
	tis620
	hexstring
)

type iso8583Tag struct {
	name       string
	isMti      bool
	field      int
	length     int
	lenEncode  encodeBase
	valEncode  encodeBase
	fieldType  iso8583FieldType
	codePage   codepageType
	bitmapSize int
}

type fixedwidthTag struct {
	name     string
	field    int
	length   int
	codePage codepageType
}

func loadTag(v reflect.Value) map[string]*iso8583Tag {
	mp := make(map[string]*iso8583Tag)
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		t, err := parseIso8583Tag(f)
		if err != nil {
			continue
		}
		mp[f.Name] = &t
	}
	return mp
}

func loadFixedwidthTag(v reflect.Value) map[string]*fixedwidthTag {
	mp := make(map[string]*fixedwidthTag)
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		t, err := parseFixedLengthTag(f)
		if err != nil {
			continue
		}
		mp[f.Name] = &t
	}
	return mp
}

func parseFixedLengthTag(f reflect.StructField) (t fixedwidthTag, err error) {
	t.name = f.Name
	if t.field, err = strconv.Atoi(f.Tag.Get(fixedFieldWord)); err != nil {
		return
	}
	if t.length, err = strconv.Atoi(f.Tag.Get(fixedLengthWord)); err != nil {
		return
	}
	t.codePage, err = parseCodepage(f.Tag.Get(fixedCodepageWord))
	return
}

func parseIso8583Tag(f reflect.StructField) (t iso8583Tag, err error) {
	var parseErr error
	t.name = f.Name
	if strings.ToLower(f.Name) == mtiWord {
		t.isMti = true
		t.valEncode = parseEncode(f.Tag.Get(encodeWord))
		return
	}
	if t.field, err = strconv.Atoi(f.Tag.Get(fieldWord)); err != nil {
		err = fmt.Errorf("field number must be specified")
		return
	}
	if t.length, parseErr = strconv.Atoi(f.Tag.Get(lengthWord)); parseErr != nil {
		t.length = -1
	}
	if t.bitmapSize, parseErr = strconv.Atoi(f.Tag.Get(bitmapsizeWord)); parseErr != nil {
		t.bitmapSize = 0
	}
	if raw := f.Tag.Get(encodeWord); raw != "" {
		enc := strings.Split(raw, ",")
		if len(enc) == 2 {
			t.lenEncode = parseEncode(enc[0])
			t.valEncode = parseEncode(enc[1])
		} else {
			t.lenEncode = ascii
			t.valEncode = parseEncode(enc[0])
		}
	} else {
		t.lenEncode = ascii
		t.valEncode = ascii
	}
	if t.fieldType, err = parseType(f.Tag.Get(typeWord)); err != nil {
		err = fmt.Errorf("type must be specified")
		return
	}
	t.codePage, err = parseCodepage(f.Tag.Get(codepageWord))
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
	case "lllvar":
		return lllvar, nil
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
	case "hexstring":
		return hexstring, nil
	}
	return -1, fmt.Errorf("Unsupport codepage %s", s)
}

func (c codepageType) value() string {
	switch c {
	case hexstring:
		return "hexstring"
	case tis620:
		return "TIS-620"
	case windows874:
		return "Windows-874"
	default:
		return ""
	}
}

func (e encodeBase) value() string {
	switch e {
	case ascii:
		return "ascii"
	case bcd:
		return "bcd"
	case rbcd:
		return "rbcd"
	default:
		return ""
	}
}

func (t iso8583FieldType) value() string {
	switch t {
	case numeric:
		return "numeric"
	case alpha:
		return "alpha"
	case binary:
		return "binary"
	case llvar:
		return "llvar"
	case lllvar:
		return "lllvar"
	default:
		return fmt.Sprintf("unknown type %v", t)
	}
}
