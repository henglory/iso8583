package iso8583

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

func nilEncoder(v reflect.Value, t tag) ([]byte, error) {
	return nil, nil
}

func nilInnerEncoder(v reflect.Value, t innerTag) ([]byte, error) {
	return nil, nil
}

func unknownEncoder(t reflect.Type) valueEncoder {
	return func(v reflect.Value, tg tag) ([]byte, error) {
		return nil, fmt.Errorf("Type %s is not support for this library value(%#v)", t.Kind(), v)
	}
}

func unknownInnerEncoder(t reflect.Type) valueInnerEncoder {
	return func(v reflect.Value, tg innerTag) ([]byte, error) {
		return nil, fmt.Errorf("Type %s is not support for this library value(%#v)", t.Kind(), v)
	}
}

func initEncoder(v reflect.Value) ([]byte, error) {
	var mti []byte
	// var secondBitmap bool
	mp := make(map[int]([]byte))
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		t, err := parseTag(f)
		if err != nil {
			continue
		}
		if t.isMti {
			mti, err = encodeMti(v.Field(i), t)
			if err != nil {
				return nil, err
			}
			continue
		}
		// if t.isSecondBitmap {
		// 	secondBitmap, err = encodeSecondBitmap(v.Field(i))
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	continue
		// }
		b, err := primaryEncoder(f.Type, t)(v.Field(i), t)
		if err != nil {
			return nil, err
		}
		mp[t.field] = b
	}

	if len(mti) == 0 {
		return nil, fmt.Errorf("Require MTI field")
	}

	return aggregateVal(mp, mti)
}

// func aggregateVal(mp map[int]([]byte), mti []byte, secondBitmap bool) ([]byte, error) {
// 	var ret []byte
// 	ret = append(ret, mti...)
// 	byteNum := 8
// 	if secondBitmap {
// 		byteNum = 16
// 	}
// 	bitmap := make([]byte, byteNum)
// 	var data []byte
// 	for byteIdx := 0; byteIdx < byteNum; byteIdx++ {
// 		for bitIdx := 0; bitIdx < 8; bitIdx++ {
// 			i := byteIdx*8 + bitIdx + 1
// 			// if we need second bitmap (additional 8 bytes) - set first bit in first bitmap
// 			// secondary bit map is the first val
// 			if secondBitmap && i == 1 {
// 				step := uint(7 - bitIdx)
// 				bitmap[byteIdx] |= (0x01 << step)
// 			}
// 			if b, ok := mp[i]; ok && len(b) > 0 {
// 				// mark 1 in bitmap:
// 				step := uint(7 - bitIdx)
// 				bitmap[byteIdx] |= (0x01 << step)
// 				data = append(data, b...)
// 			}
// 		}
// 	}
// 	ret = append(ret, bitmap...)
// 	ret = append(ret, data...)
// 	return ret, nil
// }

func aggregateVal(mp map[int]([]byte), mti []byte) ([]byte, error) {
	var ret []byte
	ret = append(ret, mti...)
	//assume has only primary bitmap
	bitmap := make([]byte, 8)
	var data []byte
	var hasSecondBitmap = false
	for idx, m := range mp {
		if len(m) <= 0 {
			continue
		}
		if idx <= 0 || idx > 128 {
			return nil, fmt.Errorf("Data DE is less than 0 or more than 128")
		}
		if idx > 64 && !hasSecondBitmap {
			//add second bitmap
			bitmap = append(bitmap, make([]byte, 8)...)
			bitmap[0] |= (0x80)
			hasSecondBitmap = true
		}
		byteIdx := idx / 8
		bitIdx := (idx - 1) % 8
		step := uint(7 - bitIdx)
		bitmap[byteIdx] |= (0x01 << step)
		data = append(data, m...)
	}

	// for byteIdx := 0; byteIdx < byteNum; byteIdx++ {
	// 	for bitIdx := 0; bitIdx < 8; bitIdx++ {
	// 		i := byteIdx*8 + bitIdx + 1
	// 		// if we need second bitmap (additional 8 bytes) - set first bit in first bitmap
	// 		// secondary bit map is the first val
	// 		if secondBitmap && i == 1 {
	// 			step := uint(7 - bitIdx)
	// 			bitmap[byteIdx] |= (0x01 << step)
	// 		}
	// 		if b, ok := mp[i]; ok && len(b) > 0 {
	// 			// mark 1 in bitmap:
	// 			step := uint(7 - bitIdx)
	// 			bitmap[byteIdx] |= (0x01 << step)
	// 			data = append(data, b...)
	// 		}
	// 	}
	// }
	ret = append(ret, bitmap...)
	ret = append(ret, data...)
	return ret, nil
}

type innerData struct {
	field  int
	length int
	b      []byte
}

type innerDataSort []innerData

func (s innerDataSort) Len() int           { return len(s) }
func (s innerDataSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s innerDataSort) Less(i, j int) bool { return s[i].field < s[j].field }

func ptrEncoder(v reflect.Value, t tag) ([]byte, error) {
	if v.IsNil() {
		return nil, nil
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return structEncoder(v, t)
}

func structEncoder(v reflect.Value, t tag) ([]byte, error) {

	var dataArr []innerData
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		innerT, err := parseInnerTag(f.Tag)
		if err != nil {
			continue
		}
		b, err := newValueInnerEncoder(f.Type)(v.Field(i), innerT)
		if err != nil {
			continue
		}
		dataArr = append(dataArr, innerData{innerT.field, innerT.length, b})
	}
	b, err := aggregateInnerVal(dataArr)
	if err != nil {
		return nil, err
	}
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

func aggregateInnerVal(arr []innerData) (byt []byte, err error) {
	sort.Sort(innerDataSort(arr))
	data := bytes.NewBuffer([]byte{})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error encode %#v", r)
		}
	}()
	for _, itr := range arr {
		_, err = data.Write(itr.b)
		if err != nil {
			return
		}
	}
	byt = data.Bytes()
	return
}

type valueInnerEncoder func(v reflect.Value, t innerTag) ([]byte, error)

func newValueInnerEncoder(t reflect.Type) valueInnerEncoder {
	if t == nil {
		return nilInnerEncoder
	}
	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		return ptrInterfaceInnerEncoder
	case reflect.String:
		return stringInnerEncoder
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intInnerEncoder
	case reflect.Float64:
		return floatInnerEncoder(0, 64)
	case reflect.Float32:
		return floatInnerEncoder(0, 32)
	}
	return unknownInnerEncoder(t)
}

func sliceEncoder(v reflect.Value, t tag) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("value is not byte array")
		}
	}()
	return parseBinary(v.Bytes(), t)
}

func stringEncoder(v reflect.Value, t tag) ([]byte, error) {
	if v.String() == "" {
		return []byte{}, nil
	}
	return parseValue(v.String(), t)
}

func intEncoder(v reflect.Value, t tag) ([]byte, error) {
	if v.Int() == 0 {
		return []byte{}, nil
	}
	return parseValue(strconv.Itoa(int(v.Int())), t)
}

func floatEncoder(perc, bitSize int) valueEncoder {
	return func(v reflect.Value, t tag) ([]byte, error) {
		if v.Float() == 0 {
			return []byte{}, nil
		}
		return parseValue(strconv.FormatFloat(v.Float(), 'f', perc, bitSize), t)
	}
}

func parseBinary(b []byte, t tag) ([]byte, error) {
	var fn parseBinaryFn
	switch t.fieldType {
	case binary:
		fn = getBinaryParser(t)
	case llvar:
		fn = getLlvarBinaryParser(t)
	case lllvar:
		fn = getLllvarBinaryParser(t)
	default:
		return nil, fmt.Errorf("type is not support for parsing binary")
	}
	return fn(b)
}

func parseValue(s string, t tag) ([]byte, error) {
	var fn parseValueFn
	switch t.fieldType {
	case numeric:
		fn = getNumbericParser(t)
	case alpha:
		fn = getAlphaParser(t)
	case llvar:
		fn = getLlvarValueParser(t)
	case llnumberic:
		fn = getLlnumValueParser(t)
	case lllvar:
		fn = getLllvarValueParser(t)
	case lllnumberic:
		fn = getLllnumValueParser(t)
	default:
		return nil, fmt.Errorf("type is not supported for parsing value")
	}
	return fn(s)
}
