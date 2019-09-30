package iso8583

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func structBitmapEncoder(v reflect.Value, t tag) ([]byte, error) {

	var dataArr []innerData
	var bitmap = make([]byte, t.bitmapSize)
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		innerT, err := parseInnerTag(f.Tag)
		if err != nil {
			continue
		}
		b, err := newValueBitmapEncoder(f.Type)(v.Field(i), innerT)
		if err != nil {
			continue
		}
		bitmap, err = calculateBitmap(bitmap, innerT.field)
		if err != nil {
			return nil, err
		}
		dataArr = append(dataArr, innerData{innerT.field, innerT.length, b})
	}
	b, err := aggregateInnerVal(dataArr)
	if err != nil {
		return nil, err
	}
	b = append(bitmap, b...)
	var fn parseBinaryFn
	switch t.fieldType {
	case llvar:
		fn = getLlvarBinaryParser(t)
	case lllvar:
		fn = getLllvarBinaryParser(t)
	default:
		return nil, fmt.Errorf("type is not support for parsing binary")
	}
	return fn(b)
}

func calculateBitmap(bitmap []byte, fieldNumber int) ([]byte, error) {
	if fieldNumber < 1 {
		return nil, errors.New("field number could not be specified below 1")
	}
	slot := (fieldNumber - 1) / 8
	if slot >= len(bitmap) {
		return nil, fmt.Errorf("bitmap size could not be smaller than field number bitmapSize:%d, fieldNumber:%d", len(bitmap), fieldNumber)
	}
	bitIdx := (fieldNumber - 1) % 8
	step := uint(7 - bitIdx)
	bitmap[slot] |= (0x01 << step)
	return bitmap, nil
}

func newValueBitmapEncoder(t reflect.Type) valueInnerEncoder {
	if t == nil {
		return nilBitmapEncoder
	}
	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		return ptrInterfaceBitmapEncoder
	case reflect.String:
		return stringBitmapEncoder
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intInnerEncoder
	case reflect.Float64:
		return floatInnerEncoder(0, 64)
	case reflect.Float32:
		return floatInnerEncoder(0, 32)
	}
	return unknownInnerEncoder(t)
}

func stringBitmapEncoder(v reflect.Value, t innerTag) ([]byte, error) {
	if v.String() == "" {
		return nil, fmt.Errorf("No value for encoding field: %d", t.field)
	}
	vb := []byte(v.String())
	if cp := t.codePage.value(); cp != "" {
		vb = encodeUTF8(cp, vb)
	}
	if len(vb) < t.length {
		vb = append(vb, []byte(strings.Repeat(" ", t.length-len(vb)))...)
	} else if len(vb) > t.length {
		vb = vb[:t.length]
	}
	return vb, nil
}

func nilBitmapEncoder(v reflect.Value, t innerTag) ([]byte, error) {
	return nil, fmt.Errorf("No value for encoding field: %d", t.field)
}

func ptrInterfaceBitmapEncoder(v reflect.Value, tg innerTag) ([]byte, error) {
	if v.IsNil() {
		return nilBitmapEncoder(v, tg)
	}
	return newValueBitmapEncoder(v.Elem().Type())(v.Elem(), tg)
}
