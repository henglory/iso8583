package iso8583v2

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type fixedwidthDecoder struct {
	v  reflect.Value
	tg fixedwidthTag
}

func (f *fixedwidthDecoder) ptrDecodeFunc(data []byte) error {
	typ := f.v.Type().Elem()
	if f.v.IsNil() {
		f.v.Set(reflect.New(typ))
	}
	for f.v.Kind() == reflect.Ptr {
		f.v = reflect.Indirect(f.v)
	}

	switch f.v.Kind() {
	case reflect.String:
		return f.stringDecodeFunc(data)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return f.intDecodeFunc(data)
	case reflect.Float32:
		return f.getFloatDecoder(32)(data)
	case reflect.Float64:
		return f.getFloatDecoder(64)(data)
	}
	return f.unknownFunc(data)
}

func (f *fixedwidthDecoder) stringDecodeFunc(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("field:%s string decode failed %v", f.tg.name, r)
		}
	}()
	if c := f.tg.codePage.value(); c != "" {
		data = decodeUTF8(c, data)
	}
	data = bytes.TrimSpace(data)
	f.v.SetString(string(data))
	return
}

func (f *fixedwidthDecoder) intDecodeFunc(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("field:%s int decode failed %v", f.tg.name, r)
		}
	}()
	if len(data) < 1 {
		return
	}
	data = bytes.TrimSpace(data)
	i, err := strconv.Atoi(string(data))
	if err != nil {
		return
	}
	f.v.SetInt(int64(i))
	return
}

func (f *fixedwidthDecoder) getFloatDecoder(bitSize int) func(data []byte) error {
	return func(data []byte) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("field:%s float decode failed %v", f.tg.name, r)
			}
		}()
		if len(data) < 1 {
			return
		}
		var fval float64
		data = bytes.TrimSpace(data)
		fval, err = strconv.ParseFloat(string(data), bitSize)
		if err != nil {
			return
		}
		f.v.SetFloat(fval)
		return
	}
}

func (f *fixedwidthDecoder) unknownFunc(data []byte) error {
	return fmt.Errorf("field:%s unsupported decoder %s", f.tg.name, f.v.Type().Kind().String())
}

func getFixedwidthDecoder(v reflect.Value, tg fixedwidthTag) func(data []byte) error {
	fEnc := &fixedwidthDecoder{
		v:  v,
		tg: tg,
	}
	switch v.Kind() {
	case reflect.Ptr:
		return fEnc.ptrDecodeFunc
	case reflect.String:
		return fEnc.stringDecodeFunc
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return fEnc.intDecodeFunc
	case reflect.Float32:
		return fEnc.getFloatDecoder(32)
	case reflect.Float64:
		return fEnc.getFloatDecoder(64)
	}

	return fEnc.unknownFunc
}
