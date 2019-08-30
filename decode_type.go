package iso8583

import (
	"encoding/hex"
	"fmt"
	"reflect"
)

//leftbyte, bitmap, err
type decoderFn func(bitmap, data []byte) ([]byte, []byte, error)

func initDecoder(v reflect.Value, raw []byte) error {
	var mtiDecoder decoderFn
	var bitmapDecoder decoderFn
	var pipeline []decoderFn
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		t, err := parseTag(f)
		if err != nil {
			continue
		}
		if t.isMti {
			mtiDecoder = getMtiDecoder(v.Field(i), t)
			continue
		}
		if t.isSecondBitmap {
			bitmapDecoder = getBitmapDecoder(v.Field(i))
			continue
		}
		pipeline = append(pipeline, getValueDecoder(v.Field(i), t))
	}
	if mtiDecoder == nil || bitmapDecoder == nil {
		return fmt.Errorf("Struct decoder not supported")
	}
	pipeline = append([]decoderFn{mtiDecoder, bitmapDecoder}, pipeline...)
	var bm []byte
	var err error
	for _, exec := range pipeline {
		raw, bm, err = exec(bm, raw)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMtiDecoder(v reflect.Value, t tag) decoderFn {
	return func(bm, data []byte) (leftByte []byte, bitmap []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("Mti decoder critical error %s", hex.EncodeToString(data))
			}
		}()
		if v.Type().Kind() != reflect.String {
			err = fmt.Errorf("MTI should be string")
			return
		}

		var mtiBytes []byte
		switch t.encode {
		case ascii:
			mtiBytes = data[:4]
			leftByte = data[4:]
			v.SetString(string(mtiBytes))
		case bcd:
			mtiBytes = data[:2]
			leftByte = data[2:]
			v.SetString(string(bcd2Ascii(mtiBytes)))
		default:
			err = fmt.Errorf("Mti invalid encode")
			return
		}
		return
	}
}

func getBitmapDecoder(v reflect.Value) decoderFn {
	return func(bm, data []byte) (leftByte []byte, bitmap []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("Bitmap decoder critical error %s", hex.EncodeToString(data))
			}
		}()
		if v.Type().Kind() != reflect.Bool {
			err = fmt.Errorf("Second bitmap flag should be boolean")
		}

		byteNum := 8
		if data[0]&0x80 == 0x80 {
			v.SetBool(true)
			byteNum = 16
		} else {
			v.SetBool(false)
		}
		bitmap = data[:byteNum]
		leftByte = data[byteNum:]
		return
	}
}

func getValueDecoder(v reflect.Value, t tag) decoderFn {
	return func(bm, data []byte) (leftByte []byte, bitmap []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("Value decoder critical error %s", hex.EncodeToString(data))
			}
		}()
		maxFieldLength := len(bm) * 8
		if t.field > maxFieldLength || len(data) == 0 {
			leftByte = []byte{}
			bitmap = bm
			return
		}
		byteIndex := (t.field - 1) / 8
		bitIndex := (t.field - 1) % 8
		if (bm[byteIndex] & (0x80 >> uint(bitIndex))) != (0x80 >> uint(bitIndex)) {
			//offbit
			leftByte = data
			bitmap = bm
			return
		}

		//onbit
		bitmap = bm
		leftByte, err = getIso8583TypeDecoder(t)(v, t, data)
		return
	}
}

type iso8583TypeDecoder func(v reflect.Value, t tag, data []byte) ([]byte, error)

func getIso8583TypeDecoder(t tag) iso8583TypeDecoder {
	switch t.fieldType {
	case numeric:
		return numericDecoder
	case alpha:
		return alphaDecoder
	case binary:
		return binaryDecoder
	case llvar:
		return llvarDecoder
	case llnumberic:
		return llnumDecoder
	case lllvar:
		return lllvarDecoder
	case lllnumberic:
		return lllnumDecoder
	}
	return unknownIso8583TypeDecoder
}

func unknownIso8583TypeDecoder(v reflect.Value, t tag, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("Unknown type decoder %d", t.fieldType)
}
