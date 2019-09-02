package iso8583

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

//Marshal is used for convert iso8583 struct to byte array
func Marshal(v interface{}) ([]byte, error) {
	return encode(v)

}

func encode(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("Not suppport null pointer")
	}
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice {
		return nil, fmt.Errorf("Not support for slice")
	}
	return write(reflect.ValueOf(v))
}

func write(v reflect.Value) ([]byte, error) {
	if v.Type().Kind() != reflect.Struct {
		return nil, fmt.Errorf("Not support for not struct type %v", v)
	}
	return initEncoder(v)
}

type valueEncoder func(v reflect.Value, t tag) ([]byte, error)

func primaryEncoder(t reflect.Type) valueEncoder {
	if t == nil {
		return nilEncoder
	}
	switch t.Kind() {
	case reflect.Slice:
		return sliceEncoder
	case reflect.Struct:
		return structEncoder
	case reflect.String:
		return stringEncoder
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intEncoder
	case reflect.Float64:
		return floatEncoder(0, 64)
	case reflect.Float32:
		return floatEncoder(0, 32)
	}
	return unknownEncoder(t)
}

func encodeMti(v reflect.Value, t tag) ([]byte, error) {
	if v.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("MTI should be string")
	}
	mti := v.String()
	if mti == "" {
		return nil, fmt.Errorf("MTI is required")
	}
	if len(mti) != 4 {
		return nil, fmt.Errorf("MTI is invalid")
	}

	// check MTI, it must contain only digits
	if _, err := strconv.Atoi(mti); err != nil {
		return nil, errors.New("MTI is invalid")
	}

	switch t.encode {
	case bcd:
		return bcdEncode([]byte(mti)), nil
	default:
		return []byte(mti), nil
	}
}

func encodeSecondBitmap(v reflect.Value) (bool, error) {
	if v.Type().Kind() != reflect.Bool {
		return false, fmt.Errorf("Second Bitmap sholud be boolean")
	}
	return v.Bool(), nil
}
