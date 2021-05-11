package iso8583v2

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync/atomic"
)

type fieldEncoder struct {
	typ reflect.Type
	tg  iso8583Tag
}

type fieldEncoderFunc func(v reflect.Value) ([]byte, error)

func (f fieldEncoder) nilFunc(v reflect.Value) ([]byte, error) {
	return nil, nil
}

func (f fieldEncoder) unknownFunc(v reflect.Value) ([]byte, error) {
	return nil, fmt.Errorf("field type %s is not support for this library value(%#v)", f.typ.Kind(), v)
}

func (f fieldEncoder) sliceByteEncodeFunc(v reflect.Value) (bret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("encode failed field:%s slice supported only byte array", f.tg.name)
		}
	}()
	byts := v.Bytes()
	if len(byts) <= 0 {
		bret = []byte{}
		err = nil
		return
	}
	bret, err = f.parseValue(byts)
	return
}

func (f fieldEncoder) ptrEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return nil, nil
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		return f.sliceByteEncodeFunc(v)
	case reflect.Struct:
		return f.structEncodeFunc(v)
	case reflect.String:
		return f.stringEncodeFunc(v)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return f.intEncodeFunc(v)
	case reflect.Float32:
		return f.getFloatEncoder(0, 32)(v)
	case reflect.Float64:
		return f.getFloatEncoder(0, 64)(v)
	default:
		return nil, fmt.Errorf("pointer field:%s type is not supported %v", f.tg.name, v.Kind())
	}
}

func (f fieldEncoder) structEncodeFunc(v reflect.Value) ([]byte, error) {
	structKey := v.Type().PkgPath() + v.Type().Name()
	fixedTagLock.RLock()
	mpTag := fixedTagCache[structKey]
	fixedTagLock.RUnlock()
	tag, ok := mpTag.Load().(map[string]*fixedwidthTag)
	if !ok {
		var tagValue atomic.Value
		tag = loadFixedwidthTag(v)
		tagValue.Store(tag)
		fixedTagLock.Lock()
		fixedTagCache[structKey] = tagValue
		fixedTagLock.Unlock()
	}

	structByte, err := f.encodeFixedwidthWithTag(v, tag)
	if err != nil {
		return nil, err
	}
	if f.tg.fieldType != llvar && f.tg.fieldType != lllvar {
		return nil, fmt.Errorf("struct field:%s encoding will support only llvar or lllvar type %s", f.tg.name, f.tg.fieldType.value())
	}
	return f.parseValue(structByte)
}

func (f fieldEncoder) encodeFixedwidthWithTag(v reflect.Value, tag map[string]*fixedwidthTag) ([]byte, error) {
	dataMap := make(map[int]([]byte))
	for i := 0; i < v.Type().NumField(); i++ {
		subField := v.Type().Field(i)
		fixedTag := tag[subField.Name]
		if fixedTag == nil {
			continue
		}
		b, err := getFixedwidthEncoder(subField.Type, *fixedTag, f.tg.bitmapSize > 0)(v.Field(i))
		if err != nil {
			return nil, err
		}
		if b != nil {
			dataMap[fixedTag.field] = b
		}
	}
	return f.encodeFixedwidthStructValue(dataMap)
}

func (f fieldEncoder) encodeFixedwidthStructValue(dataMap map[int]([]byte)) ([]byte, error) {
	var hasBitmap bool
	var bitmap []byte
	var data []byte
	if f.tg.bitmapSize > 0 {
		hasBitmap = true
		bitmap = make([]byte, f.tg.bitmapSize)
	}
	var keys []int
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, idx := range keys {
		m := dataMap[idx]
		if idx < 1 {
			return nil, fmt.Errorf("field:%s has sub field below 1", f.tg.name)
		}
		if len(m) <= 0 {
			continue
		}
		if hasBitmap {
			slot := (idx - 1) / 8
			if slot >= len(bitmap) {
				return nil, fmt.Errorf("field:%s is using bitmap(%d) but field count is moreover %d", f.tg.name, len(bitmap), idx)
			}
			bitIdx := (idx - 1) % 8
			step := uint(7 - bitIdx)
			bitmap[slot] |= (0x01 << step)
		}
		data = append(data, m...)
	}
	if hasBitmap {
		data = append(bitmap, data...)
	}
	return data, nil
}

func (f fieldEncoder) stringEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.String() == "" {
		return []byte{}, nil
	}
	return f.parseValue([]byte(v.String()))
}

func (f fieldEncoder) intEncodeFunc(v reflect.Value) ([]byte, error) {
	if v.Int() == 0 {
		return []byte{}, nil
	}
	return f.parseValue([]byte(strconv.Itoa(int(v.Int()))))
}

func (f fieldEncoder) getFloatEncoder(perc, bitsize int) fieldEncoderFunc {
	return func(v reflect.Value) ([]byte, error) {
		if v.Float() == 0 {
			return []byte{}, nil
		}
		return f.parseValue([]byte(strconv.FormatFloat(v.Float(), 'f', perc, bitsize)))
	}
}

func (f fieldEncoder) parseValue(b []byte) ([]byte, error) {
	switch f.tg.fieldType {
	case numeric:
		return f.numericParse(b)
	case alpha:
		return f.alphaParse(b)
	case binary:
		return f.binaryParse(b)
	case llvar:
		return f.llvarParse(b)
	case lllvar:
		return f.lllvarParse(b)
	}
	return nil, fmt.Errorf("Field:%s type is invalid(%s)", f.tg.name, f.tg.fieldType.value())
}

func getFieldEncoder(typ reflect.Type, tg iso8583Tag) fieldEncoderFunc {
	fEnc := fieldEncoder{
		typ: typ,
		tg:  tg,
	}
	if typ == nil {
		return fEnc.nilFunc
	}

	switch typ.Kind() {
	case reflect.Slice:
		return fEnc.sliceByteEncodeFunc
	case reflect.Ptr:
		return fEnc.ptrEncodeFunc
	case reflect.Struct:
		return fEnc.structEncodeFunc
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
