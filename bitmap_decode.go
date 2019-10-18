package iso8583

import (
	"errors"
	"fmt"
	"reflect"
)

func bitmapPtrDecode(v reflect.Value, t tag, data []byte) error {
	if v.IsNil() {
		return nil
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return bitmapStructDecode(v, t, data)
}

func bitmapStructDecode(v reflect.Value, t tag, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error, bitmap struct decode")
		}
	}()
	idx := t.bitmapSize
	bitmap := data[:t.bitmapSize]
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		innerT, err := parseInnerTag(f.Tag)
		if err != nil {
			err = nil
			continue
		}
		isOn, err := bitIsOn(bitmap, innerT.field)
		if err != nil {
			break
		}
		if !isOn {
			continue
		}
		err = newValueInnerDecoder(f.Type)(v.Field(i), data[idx:idx+innerT.length], innerT.codePage)
		if err != nil {
			break
		}
		idx = idx + innerT.length
	}
	return
}

func bitIsOn(bitmap []byte, fieldNumber int) (bool, error) {
	if fieldNumber < 1 {
		return false, errors.New("field number could not be specified below 1")
	}
	slot := (fieldNumber - 1) / 8
	if slot >= len(bitmap) {
		return false, fmt.Errorf("bitmap size could not be smaller than field number bitmapSize:%d, fieldNumber:%d", len(bitmap), fieldNumber)
	}
	bitIdx := (fieldNumber - 1) % 8
	if (bitmap[slot] & (0x80 >> uint(bitIdx))) == (0x80 >> uint(bitIdx)) {
		return true, nil
	}
	return false, nil
}
