package iso8583

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/henglory/iso8583/charset"
)

func ptrInterfaceInnerEncoder(v reflect.Value, tg innerTag) ([]byte, error) {
	if v.IsNil() {
		return nilInnerEncoder(v, tg)
	}
	return newValueInnerEncoder(v.Elem().Type())(v.Elem(), tg)
}

func stringInnerEncoder(v reflect.Value, t innerTag) ([]byte, error) {
	vb := []byte(v.String())
	if cp := t.codePage.value(); cp != "" {
		vb = charset.EncodeUTF8(cp, vb)
	}
	if len(vb) < t.length {
		vb = append(vb, []byte(strings.Repeat(" ", t.length-len(vb)))...)
	} else if len(vb) > t.length {
		vb = vb[:t.length]
	}
	return vb, nil
}

func intInnerEncoder(v reflect.Value, t innerTag) ([]byte, error) {
	vb := []byte(strconv.Itoa(int(v.Int())))
	if len(vb) < t.length {
		vb = append([]byte(strings.Repeat("0", t.length-len(vb))), vb...)
	} else if len(vb) > t.length {
		vb = vb[:t.length]
	}
	return vb, nil
}

func floatInnerEncoder(perc, bitSize int) valueInnerEncoder {
	return func(v reflect.Value, t innerTag) ([]byte, error) {
		vb := []byte(strconv.FormatFloat(v.Float(), 'f', perc, bitSize))
		if len(vb) < t.length {
			vb = append([]byte(strings.Repeat("0", t.length-len(vb))), vb...)
		} else if len(vb) > t.length {
			vb = vb[:t.length]
		}
		return vb, nil
	}
}
