package iso8583v2

import (
	"fmt"
	"strings"
)

func (f fieldEncoder) numericParse(b []byte) ([]byte, error) {
	if f.tg.length <= 0 {
		return nil, fmt.Errorf("numeric field:%s length must be specified", f.tg.name)
	}
	if f.tg.valEncode == rbcd &&
		len(b) == (f.tg.length+1) &&
		string(b[0:1]) == "0" {
		b = b[1:len(b)]
	}
	if len(b) > f.tg.length {
		return nil, fmt.Errorf("numeric field:%s length (%d) is larger than defined length (%d)", f.tg.name, len(b), f.tg.length)
	}
	if len(b) < f.tg.length {
		b = append([]byte(strings.Repeat("0", f.tg.length-len(b))), b...)
	}
	switch f.tg.valEncode {
	case bcd:
		return lbcdEncode(b)
	case rbcd:
		return rbcdEncode(b)
	case ascii:
		return b, nil
	//default will be ascii
	default:
		return b, nil
	}
}

func (f fieldEncoder) alphaParse(b []byte) ([]byte, error) {
	if f.tg.length <= 0 {
		return nil, fmt.Errorf("alpha field:%s length must be specified", f.tg.name)
	}
	if cp := f.tg.codePage.value(); cp != "" {
		b = encodeUTF8(cp, b)
	}
	if len(b) > f.tg.length {
		return nil, fmt.Errorf("alpha field:%s data %s length (%d) is larger than defined length (%d)", f.tg.name, string(b), len(b), f.tg.length)
	}
	if len(b) < f.tg.length {
		b = append(b, []byte(strings.Repeat(" ", f.tg.length-len(b)))...)
	}
	return b, nil
}

func (f fieldEncoder) binaryParse(b []byte) ([]byte, error) {
	if f.tg.length <= 0 {
		return nil, fmt.Errorf("binary field:%s length must be specified", f.tg.name)
	}
	if len(b) > f.tg.length {
		return nil, fmt.Errorf("binary field:%s length (%d) is larger than defined length (%d)", f.tg.name, len(b), f.tg.length)
	}
	if len(b) < f.tg.length {
		return append(b, make([]byte, f.tg.length-len(b))...), nil
	}
	return b, nil
}

func (f fieldEncoder) llvarParse(b []byte) ([]byte, error) {
	if cp := f.tg.codePage.value(); cp != "" {
		b = encodeUTF8(cp, b)
	}
	if f.tg.length != -1 && len(b) > f.tg.length {
		return nil, fmt.Errorf("llvar field:%s length is defined but value(%d) is larger than defined(%d)", f.tg.name, len(b), f.tg.length)
	}
	if f.tg.valEncode != ascii {
		return nil, fmt.Errorf("llvar field:%s value encode must be ascii", f.tg.name)
	}
	contentLen := []byte(fmt.Sprintf("%02d", len(b)))
	var lenVal []byte
	switch f.tg.lenEncode {
	case ascii:
		lenVal = contentLen
		if len(lenVal) > 2 {
			return nil, fmt.Errorf("llvar field:%s ascii length value is invalid(%d)", f.tg.name, len(lenVal))
		}
	case rbcd:
		fallthrough
	case bcd:
		var err error
		lenVal, err = rbcdEncode(contentLen)
		if err != nil {
			return nil, fmt.Errorf("llvar field:%s rbcd encode failed %s", f.tg.name, err.Error())
		}
		if len(lenVal) > 1 || len(contentLen) > 3 {
			return nil, fmt.Errorf("llvar field:%s bcd length value is invalid(%d) content length(%d)", f.tg.name, len(lenVal), len(contentLen))
		}
	default:
		return nil, fmt.Errorf("llvar field:%s length encode is not valid %s", f.tg.name, f.tg.lenEncode.value())
	}
	return append(lenVal, b...), nil
}

func (f fieldEncoder) lllvarParse(b []byte) ([]byte, error) {
	if cp := f.tg.codePage.value(); cp != "" {
		b = encodeUTF8(cp, b)
	}
	if f.tg.length != -1 && len(b) > f.tg.length {
		return nil, fmt.Errorf("lllvar field:%s length is defined but value(%d) is larger than defined(%d)", f.tg.name, len(b), f.tg.length)
	}
	if f.tg.valEncode != ascii {
		return nil, fmt.Errorf("lllvar field:%s value encode must be ascii", f.tg.name)
	}
	contentLen := []byte(fmt.Sprintf("%03d", len(b)))
	var lenVal []byte
	switch f.tg.lenEncode {
	case ascii:
		lenVal = contentLen
		if len(lenVal) > 3 {
			return nil, fmt.Errorf("lllvar field:%s ascii length value is invalid(%d)", f.tg.name, len(lenVal))
		}
	case rbcd:
		fallthrough
	case bcd:
		var err error
		lenVal, err = rbcdEncode(contentLen)
		if err != nil {
			return nil, fmt.Errorf("lllvar field:%s rbcd encode failed %s", f.tg.name, err.Error())
		}
		if len(lenVal) > 2 || len(contentLen) > 3 {
			return nil, fmt.Errorf("lllvar field:%s bcd length value is invalid(%d) content length(%d)", f.tg.name, len(lenVal), len(contentLen))
		}
	default:
		return nil, fmt.Errorf("lllvar field:%s length encode is not valid %s", f.tg.name, f.tg.lenEncode.value())
	}
	return append(lenVal, b...), nil
}
