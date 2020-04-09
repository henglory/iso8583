package iso8583v2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type fixedwidthEncoder struct {
	typ reflect.Type
	tg  fixedwidthTag
}

type fixedwidthEncoderFunc func(v reflect.Value) ([]byte, error)

func (f fixedwidthEncoder) nilFunc(v reflect.Value) ([]byte, error) {
	return nil, nil
}

func (f fixedwidthEncoder) unknownFunc(v reflect.Value) ([]byte, error) {
	return nil, fmt.Errorf("fixedwidth type %s is not support for this library value(%#v)", f.typ.Kind(), v)
}

func (f fixedwidthEncoder) ptrEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return nil, nil
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return f.stringEncodeFunc(v)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return f.intEncodeFunc(v)
	case reflect.Float32:
		return f.getFloatEncoder(0, 32)(v)
	case reflect.Float64:
		return f.getFloatEncoder(0, 64)(v)
	default:
		return nil, fmt.Errorf("fixedwidth pointer field:%s type is not supported %v", f.tg.name, v.Kind())
	}
}

func (f fixedwidthEncoder) stringEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.String() == "" {
		return []byte{}, nil
	}
	return f.parseStringValue([]byte(v.String()))
}

func (f fixedwidthEncoder) intEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.Int() == 0 {
		return []byte{}, nil
	}
	return f.parseNumericValue([]byte(strconv.Itoa(int(v.Int()))))
}

func (f fixedwidthEncoder) getFloatEncoder(perc, bitsize int) fixedwidthEncoderFunc {
	return func(v reflect.Value) ([]byte, error) {
		if v.Float() == 0 {
			return []byte{}, nil
		}
		return f.parseNumericValue([]byte(strconv.FormatFloat(v.Float(), 'f', perc, bitsize)))
	}
}

func (f fixedwidthEncoder) parseNumericValue(b []byte) ([]byte, error) {
	if cp := f.tg.codePage.value(); cp != "" {
		b = encodeUTF8(cp, b)
	}
	if len(b) > f.tg.length {
		return nil, fmt.Errorf("fixed width field:%s numberic is larger than configure %d, actual %d", f.tg.name, f.tg.length, len(b))
	}
	if len(b) < f.tg.length {
		b = append([]byte(strings.Repeat("0", f.tg.length-len(b))), b...)
	}
	return b, nil
}

func (f fixedwidthEncoder) parseStringValue(b []byte) ([]byte, error) {
	if cp := f.tg.codePage.value(); cp != "" {
		b = encodeUTF8(cp, b)
	}
	if len(b) > f.tg.length {
		b = b[:f.tg.length]
	}
	if len(b) < f.tg.length {
		b = append(b, []byte(strings.Repeat(" ", f.tg.length-len(b)))...)
	}
	return b, nil
}

func getFixedwidthEncoder(typ reflect.Type, tg fixedwidthTag) fixedwidthEncoderFunc {
	fEnc := fixedwidthEncoder{
		typ: typ,
		tg:  tg,
	}
	if typ == nil {
		return fEnc.nilFunc
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return fEnc.ptrEncodeFunc
	case reflect.String:
		return fEnc.stringEncodeFunc
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return fEnc.intEncodeFunc
	case reflect.Float32:
		return fEnc.getFloatEncoder(0, 32)
	case reflect.Float64:
		return fEnc.getFloatEncoder(0, 64)
	}

	return fEnc.unknownFunc
}
