package iso8583

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type encoder struct {
	w *bufio.Writer
}

//Marshal is used for convert iso8583 struct to byte array
func Marshal(v interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	err := newEncoder(buff).encode(v)
	if err != nil {
		return nil, err
	}
	msg := buff.Bytes()
	// ADD header to ISO Message
	// header := []byte(fmt.Sprintf("ISO00%04d", len(msg)))
	// msg = append(header, msg...)
	// dataLen, err := hex.DecodeString(fmt.Sprintf("%04x", len(msg)))
	// if err != nil {
	// 	return nil, err
	// }
	// msg = append(dataLen, msg...)
	return msg, nil
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		bufio.NewWriter(w),
	}
}

func (e *encoder) encode(v interface{}) error {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice {
		return fmt.Errorf("Not support for slice")
	}
	err := e.write(reflect.ValueOf(v))
	if err != nil {
		return err
	}
	return e.w.Flush()
}

func (e *encoder) write(v reflect.Value) error {
	if v.Type().Kind() != reflect.Struct {
		return fmt.Errorf("Not support for not struct type %v", v)
	}
	b, err := initEncoder(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
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
