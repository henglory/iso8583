package iso8583

import (
	"fmt"
	"strings"
)

type parseValueFn func(val string) ([]byte, error)

type parseBinaryFn func(b []byte) ([]byte, error)

func getNumbericParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		if t.length == -1 {
			return nil, fmt.Errorf("In numberic, length must be specified")
		}
		vb := []byte(val)

		if t.encode == rbcd && len(vb) == (t.length+1) && string(vb[0:1]) == "0" {
			vb = vb[1:len(vb)]
		}
		if len(vb) > t.length {
			return nil, fmt.Errorf("In numberic, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		if len(vb) < t.length {
			vb = append([]byte(strings.Repeat("0", t.length-len(vb))), vb...)
		}
		switch t.encode {
		case bcd:
			return lbcdEncode(vb), nil
		case rbcd:
			return rbcdEncode(vb), nil
		case ascii:
			return vb, nil
		}
		return nil, fmt.Errorf("In numberic, should not be executed")
	}
}

func getAlphaParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		if t.length == -1 {
			return nil, fmt.Errorf("In alpha, length must be specified")
		}
		vb := []byte(val)
		if cp := t.codePage.value(); cp != "" {
			vb = encodeUTF8(cp, vb)
		}
		if len(vb) > t.length {
			return nil, fmt.Errorf("In alpha, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		if len(vb) < t.length {
			vb = append(vb, []byte(strings.Repeat(" ", t.length-len(vb)))...)
		}
		return vb, nil
	}
}

func getBinaryParser(t tag) parseBinaryFn {
	return func(val []byte) ([]byte, error) {
		if t.length == -1 {
			return nil, fmt.Errorf("In binary, length must be specified")
		}
		if t.fieldType != binary {
			return nil, fmt.Errorf("In binary, field type should be binary and byte array")
		}
		if len(val) > t.length {
			return nil, fmt.Errorf("In binary, length (%d) is larger than defined length (%d)", len(val), t.length)
		}
		if len(val) < t.length {
			return append(val, make([]byte, t.length-len(val))...), nil
		}
		return val, nil
	}
}

func getLlvarValueParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		vb := []byte(val)
		if cp := t.codePage.value(); cp != "" {
			vb = encodeUTF8(cp, vb)
		}
		return getLlvarBinaryParser(t)(vb)
	}
}

func getLlvarBinaryParser(t tag) parseBinaryFn {
	return func(vb []byte) ([]byte, error) {
		if t.length != -1 && len(vb) > t.length {
			return nil, fmt.Errorf("In llvar, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		if t.encode != ascii {
			return nil, fmt.Errorf("In llvar, encode is not ascii")
		}

		lenStr := fmt.Sprintf("%02d", len(vb))
		contentLen := []byte(lenStr)
		var lenVal []byte
		switch t.lenEncode {
		case ascii:
			lenVal = contentLen
			if len(lenVal) > 2 {
				return nil, fmt.Errorf("In llvar, ascii length value is invalid: %d", len(lenVal))
			}
		case rbcd:
			fallthrough
		case bcd:
			lenVal = rbcdEncode(contentLen)
			if len(lenVal) > 1 || len(contentLen) > 3 {
				return nil, fmt.Errorf("In llvar, bcd length value is invalid: %d", len(lenVal))
			}
		default:
			return nil, fmt.Errorf("In llvar, encode type is not valid")
		}
		return append(lenVal, vb...), nil
	}
}

func getLlnumValueParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		vb := []byte(val)
		if t.length != -1 && len(vb) > t.length {
			return nil, fmt.Errorf("In llnumeric, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		switch t.encode {
		case ascii:
		case bcd:
			vb = lbcdEncode(vb)
		case rbcd:
			vb = rbcdEncode(vb)
		default:
			return nil, fmt.Errorf("In llnumeric, encode type is not valid")
		}

		lenStr := fmt.Sprintf("%02d", len(vb)) // length of digital characters
		contentLen := []byte(lenStr)
		var lenVal []byte
		switch t.lenEncode {
		case ascii:
			lenVal = contentLen
			if len(lenVal) > 2 {
				return nil, fmt.Errorf("In llnumberic, ascii length value is invalid: %d", len(lenVal))
			}
		case rbcd:
			fallthrough
		case bcd:
			lenVal = rbcdEncode(contentLen)
			if len(lenVal) > 1 || len(contentLen) > 3 {
				return nil, fmt.Errorf("In llnumberic, bcd length value is invalid: %d", len(lenVal))
			}
		default:
			return nil, fmt.Errorf("In llnumeric, encode type is not valid")
		}
		return append(lenVal, vb...), nil
	}
}

func getLllvarValueParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		vb := []byte(val)
		if cp := t.codePage.value(); cp != "" {
			vb = encodeUTF8(cp, vb)
		}
		return getLllvarBinaryParser(t)(vb)
	}
}

func getLllvarBinaryParser(t tag) parseBinaryFn {
	return func(vb []byte) ([]byte, error) {
		if t.length != -1 && len(vb) > t.length {
			return nil, fmt.Errorf("In lllvar, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		if t.encode != ascii {
			return nil, fmt.Errorf("In lllvar, encode is not ascii")
		}

		lenStr := fmt.Sprintf("%03d", len(vb))
		contentLen := []byte(lenStr)
		var lenVal []byte
		switch t.lenEncode {
		case ascii:
			lenVal = contentLen
			if len(lenVal) > 3 {
				return nil, fmt.Errorf("In lllvar, ascii length value is invalid: %d", len(lenVal))
			}
		case rbcd:
			fallthrough
		case bcd:
			lenVal = rbcdEncode(contentLen)
			if len(lenVal) > 2 || len(contentLen) > 3 {
				return nil, fmt.Errorf("In lllvar, bcd length value is invalid: %d", len(lenVal))
			}
		default:
			return nil, fmt.Errorf("In lllvar, encode type is not valid")
		}
		return append(lenVal, vb...), nil
	}
}

func getLllnumValueParser(t tag) parseValueFn {
	return func(val string) ([]byte, error) {
		vb := []byte(val)
		if t.length != -1 && len(vb) > t.length {
			return nil, fmt.Errorf("In llnumeric, length (%d) is larger than defined length (%d)", len(vb), t.length)
		}
		switch t.encode {
		case ascii:
		case bcd:
			vb = lbcdEncode(vb)
		case rbcd:
			vb = rbcdEncode(vb)
		default:
			return nil, fmt.Errorf("In llnumeric, encode type is not valid")
		}

		lenStr := fmt.Sprintf("%03d", len(vb)) // length of digital characters
		contentLen := []byte(lenStr)
		var lenVal []byte
		switch t.lenEncode {
		case ascii:
			lenVal = contentLen
			if len(lenVal) > 3 {
				return nil, fmt.Errorf("In llnumberic, ascii length value is invalid: %d", len(lenVal))
			}
		case rbcd:
			fallthrough
		case bcd:
			lenVal = rbcdEncode(contentLen)
			if len(lenVal) > 2 || len(contentLen) > 3 {
				return nil, fmt.Errorf("In llnumberic, bcd length value is invalid: %d", len(lenVal))
			}
		default:
			return nil, fmt.Errorf("In llnumeric, encode type is not valid")
		}
		return append(lenVal, vb...), nil
	}
}
