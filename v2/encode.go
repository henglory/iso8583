package iso8583v2

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync/atomic"
)

func Marshal(v interface{}) ([]byte, error) {
	val, err := validateEncode(v)
	if err != nil {
		return nil, fmt.Errorf("validate failed: %s", err.Error())
	}
	//lookup for tag in cache
	structKey := val.Type().PkgPath() + val.Type().Name()
	mpTag := tagCache[structKey]
	tag, ok := mpTag.Load().(map[string]*iso8583Tag)
	if !ok {
		var tagValue atomic.Value
		tag = loadTag(val)
		tagValue.Store(tag)
		tagCache[structKey] = tagValue
	}
	//////////////////////////
	return encodeIso8583wthTag(val, tag)
}

func encodeIso8583wthTag(v reflect.Value, tag map[string]*iso8583Tag) ([]byte, error) {
	var mti []byte
	var err error
	dataMap := make(map[int][]byte)
	for i := 0; i < v.Type().NumField(); i++ {
		field := v.Type().Field(i)
		isoTag := tag[field.Name]
		if isoTag == nil {
			continue
		}
		if isoTag.isMti {
			mti, err = encodeMti(v.Field(i), *isoTag)
			if err != nil {
				return nil, fmt.Errorf("Encode %v failed because mti %v", v, v.Field(i))
			}
			continue
		}
		b, err := getFieldEncoder(field.Type, *isoTag)(v.Field(i))
		if err != nil {
			return nil, err
		}
		if b != nil {
			dataMap[isoTag.field] = b
		}
	}
	if len(mti) == 0 {
		return nil, fmt.Errorf("Encode %v failed because mti is required", v)
	}
	return encodeStructValue(dataMap, mti)
}

func encodeStructValue(dataMap map[int]([]byte), mti []byte) ([]byte, error) {
	var ret []byte
	ret = append(ret, mti...)
	bitmap := make([]byte, 8)
	var data []byte
	var hasSecondBitmap = false

	var keys []int
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, idx := range keys {
		m := dataMap[idx]
		if len(m) <= 0 {
			continue
		}
		if idx <= 0 || idx > 128 {
			return nil, fmt.Errorf("Accepted only primary and secondary bitmap idx > 0 and idx <= 128")
		}
		if idx > 64 && !hasSecondBitmap {
			//add second bitmap
			bitmap = append(bitmap, make([]byte, 8)...)
			bitmap[0] |= (0x80)
			hasSecondBitmap = true
		}
		byteIdx := (idx - 1) / 8
		bitIdx := (idx - 1) % 8
		step := uint(7 - bitIdx)
		bitmap[byteIdx] |= (0x01 << step)
		data = append(data, m...)
	}
	ret = append(ret, bitmap...)
	ret = append(ret, data...)
	return ret, nil
}

func encodeMti(v reflect.Value, t iso8583Tag) ([]byte, error) {
	if v.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("MTI type must be string")
	}
	mti := v.String()
	if mti == "" {
		return nil, fmt.Errorf("MTI value must be defined")
	}
	if len(mti) != 4 {
		return nil, fmt.Errorf("MTI length must be onnly 4")
	}

	// check MTI, it must contain only digits
	if _, err := strconv.Atoi(mti); err != nil {
		return nil, errors.New("MTI must be only numeric")
	}

	switch t.valEncode {
	case bcd:
		return bcdEncode([]byte(mti))
	default:
		return []byte(mti), nil
	}
}

//Receive interface and validate
func validateEncode(v interface{}) (reflect.Value, error) {
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
		return reflect.ValueOf(v), fmt.Errorf("null pointer %v", v)
	}
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice {
		return val, fmt.Errorf("not support for slice %v", val)
	}
	if val.Type().Kind() != reflect.Struct {
		return val, fmt.Errorf("not support for not struct type %v", val)
	}
	return val, nil
}
