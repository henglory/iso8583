package iso8583v2

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
)

func (f *fieldDecoder) numericDecode(data []byte) ([]byte, error) {
	val, leftByte, err := f.getValueEncoderFn()(data)
	if err != nil {
		return nil, fmt.Errorf("numeric decode %s", err.Error())
	}
	err = f.loadValue(val)
	if err != nil {
		return nil, fmt.Errorf("numeric decode %s", err.Error())
	}
	return leftByte, nil
}

func (f *fieldDecoder) alphaDecode(data []byte) ([]byte, error) {
	val, leftByte, err := f.getValueEncoderFn()(data)
	if err != nil {
		return nil, fmt.Errorf("alpha decode %s", err.Error())
	}
	err = f.loadValue(val)
	if err != nil {
		return nil, fmt.Errorf("alpha decode %s", err.Error())
	}
	return leftByte, nil
}

func (f *fieldDecoder) binaryDecode(data []byte) (leftByte []byte, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("field:%s load value failed cannot set value %v", f.tg.name, r)
		}
	}()
	var val []byte
	val, leftByte, err = f.getValueEncoderFn()(data)
	if err != nil {
		err = fmt.Errorf("binary decode %s", err.Error())
		return
	}
	f.v.SetBytes(val)
	return
}

func (f *fieldDecoder) llvarDecode(data []byte) (leftByte []byte, err error) {

	var contentLen int
	switch f.tg.lenEncode {
	case ascii:
		if len(data) < 2 {
			err = fmt.Errorf("llvar ascii data length is too small")
			return
		}
		contentLen, err = strconv.Atoi(string(data[:2]))
		data = data[2:]
		if err != nil {
			err = fmt.Errorf("parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		if len(data) < 1 {
			err = fmt.Errorf("llvar bcd data length is too small")
			return
		}
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:1], 2)))
		data = data[1:]
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("llvar, length encoder is invalid")
		return
	}
	val := data[:contentLen]
	leftByte = data[contentLen:]
	err = f.loadValue(val)
	return
}

func (f *fieldDecoder) lllvarDecode(data []byte) (leftByte []byte, err error) {
	var contentLen int
	switch f.tg.lenEncode {
	case ascii:
		if len(data) < 3 {
			err = fmt.Errorf("lllvar ascii data length is too small")
			return
		}
		contentLen, err = strconv.Atoi(string(data[:3]))
		data = data[3:]
		if err != nil {
			err = fmt.Errorf("parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		if len(data) < 2 {
			err = fmt.Errorf("lllvar bcd data length is too small")
			return
		}
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:2], 3)))
		data = data[2:]
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("lllvar, length encoder is invalid")
		return
	}
	val := data[:contentLen]
	leftByte = data[contentLen:]
	err = f.loadValue(val)
	return
}

//////////////////////////////////////////////////////////////
func (f *fieldDecoder) loadPointer(val []byte) (err error) {
	for f.v.Kind() == reflect.Ptr {
		f.v = reflect.Indirect(f.v)
	}
	return f.loadValue(val)
}

func (f *fieldDecoder) loadStruct(val []byte) (err error) {
	structKey := f.v.Type().PkgPath() + f.v.Type().Name()
	mpTag := fixedTagCache[structKey]
	tag, ok := mpTag.Load().(map[string]*fixedwidthTag)
	if !ok {
		var tagValue atomic.Value
		tag = loadFixedwidthTag(f.v)
		tagValue.Store(tag)
		tagCache[structKey] = tagValue
	}
	err = f.loadStructSubFieldWithTag(val, tag)
	return
}

func (f *fieldDecoder) loadStructSubFieldWithTag(val []byte, tag map[string]*fixedwidthTag) error {
	hasBitmap := f.tg.bitmapSize > 0
	var bitmap []byte
	if f.tg.bitmapSize > len(val) {
		return fmt.Errorf("field:%s bitmap size(%d) is too big than data(%d)", f.tg.name, f.tg.bitmapSize, len(val))
	}

	var idx int
	if hasBitmap {
		bitmap = val[:f.tg.bitmapSize]
		idx = f.tg.bitmapSize
	}

	for i := 0; i < f.v.Type().NumField(); i++ {
		subField := f.v.Type().Field(i)
		fixedTag := tag[subField.Name]
		if fixedTag == nil {
			continue
		}
		if hasBitmap {
			isOn, err := isBitOn(bitmap, fixedTag.field)
			if err != nil {
				return fmt.Errorf("field%s subfield:%s %s", f.tg.name, fixedTag.name, err.Error())
			}
			if !isOn {
				continue
			}
		}
		if idx+fixedTag.length > len(val) {
			return fmt.Errorf("field:%s data is not enough accumulate length(%d) data(%d)", fixedTag.name, idx+fixedTag.length, len(val))
		}
		errDecode := getFixedwidthDecoder(f.v.Field(i), *fixedTag)(val[idx : idx+fixedTag.length])
		if errDecode != nil {
			return fmt.Errorf("decode failed field:%s subfield:%s %s", f.tg.name, fixedTag.name, errDecode.Error())
		}
		idx = idx + fixedTag.length
	}
	return nil
}

func isBitOn(bitmap []byte, fieldNumber int) (bool, error) {
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

func (f *fieldDecoder) loadValue(val []byte) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("field:%s load value failed cannot set value %v", f.tg.name, r)
		}
	}()
	switch f.v.Type().Kind() {
	case reflect.Ptr:
		err = f.loadPointer(val)
		return
	case reflect.Struct:
		err = f.loadStruct(val)
		return
	case reflect.Slice:
		f.v.SetBytes(val)
		return
	case reflect.String:
		if cp := f.tg.codePage.value(); cp != "" {
			val = decodeUTF8(cp, val)
		}
		f.v.SetString(string(val))
		return
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		var i int
		i, err = strconv.Atoi(string(val))
		if err != nil {
			return err
		}
		f.v.SetInt(int64(i))
		return
	case reflect.Float64:
		var fval float64
		fval, err = strconv.ParseFloat(string(val), 64)
		if err != nil {
			return
		}
		f.v.SetFloat(fval)
		return
	case reflect.Float32:
		var fval float64
		fval, err = strconv.ParseFloat(string(val), 32)
		if err != nil {
			return
		}
		f.v.SetFloat(fval)
		return
	default:
		err = fmt.Errorf("field:%s reflect type is not supported", f.tg.name)
		return
	}
}
