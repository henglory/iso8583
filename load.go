package iso8583

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
)

type decodeEncoderFn func(data []byte, length int) ([]byte, []byte, error)

func getDecodeEncoderType(e encodeBase) decodeEncoderFn {
	switch e {
	case bcd:
		return bcdDecode
	case rbcd:
		return rbcdDecode
	case ascii:
		return asciiDecode
	default:
		return unknownDecode
	}
}

func bcdDecode(data []byte, length int) ([]byte, []byte, error) {
	l := (length + 1) / 2
	if len(data) < l {
		return nil, nil, fmt.Errorf("Length data is smaller than expected")
	}
	return bcdl2Ascii(data[:l], length), data[l:], nil
}

func rbcdDecode(data []byte, length int) ([]byte, []byte, error) {
	l := (length + 1) / 2
	if len(data) < l {
		return nil, nil, fmt.Errorf("Length data is smaller than expected")
	}
	return bcdr2Ascii(data[:l], length), data[l:], nil
}

func asciiDecode(data []byte, length int) ([]byte, []byte, error) {
	if len(data) < length {
		return nil, nil, fmt.Errorf("In numberic, length data is smaller than expected")
	}
	return data[:length], data[length:], nil
}

func unknownDecode(data []byte, length int) ([]byte, []byte, error) {
	return nil, nil, fmt.Errorf("Unsupport decode")
}

func numericDecoder(v reflect.Value, t tag, data []byte) ([]byte, error) {
	if t.length == -1 {
		return nil, fmt.Errorf("In numberic, length of data should be specified")
	}
	val, leftBytes, err := getDecodeEncoderType(t.encode)(data, t.length)
	if err != nil {
		return nil, err
	}
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(val))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return nil, err
		}
		v.SetInt(int64(i))
	case reflect.Float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	default:
		return nil, fmt.Errorf("Not support for numberic type")
	}

	return leftBytes, nil
}

func alphaDecoder(v reflect.Value, t tag, data []byte) ([]byte, error) {
	if t.length == -1 {
		return nil, fmt.Errorf("In alpha, length of data should be specified")
	}

	val, leftBytes, err := getDecodeEncoderType(t.encode)(data, t.length)
	if err != nil {
		return nil, err
	}
	val = bytes.TrimRight(val, " ")
	switch v.Type().Kind() {
	case reflect.String:
		if cp := t.codePage.value(); cp != "" {
			val = decodeUTF8(cp, val)
		}
		v.SetString(string(val))
	default:
		return nil, fmt.Errorf("Not support for alpha type")
	}
	return leftBytes, nil
}

func binaryDecoder(v reflect.Value, t tag, data []byte) (leftBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("In binary, field must be byte array")
			return
		}
	}()
	if t.length == -1 {
		err = fmt.Errorf("In binary, length of data must be specified")
		return
	}
	val, leftBytes, err := getDecodeEncoderType(t.encode)(data, t.length)
	if err != nil {
		return
	}
	v.SetBytes(val)
	return
}

func llvarDecoder(v reflect.Value, t tag, data []byte) (leftBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error in llvar")
			return
		}
	}()
	var contentLen int
	switch t.lenEncode {
	case ascii:
		leftBytes = data[2:]
		contentLen, err = strconv.Atoi(string(data[:2]))
		if err != nil {
			err = fmt.Errorf("Parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		leftBytes = data[1:]
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:1], 2)))
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("In llvar, length encoder is invalid")
		return
	}
	val := leftBytes[:contentLen]
	leftBytes = leftBytes[contentLen:]
	switch v.Type().Kind() {
	case reflect.Ptr:
		if t.bitmapSize > 0 {
			err = bitmapPtrDecode(v, t, val)
			if err != nil {
				return
			}
		} else {
			err = ptrDecode(v, val)
			if err != nil {
				return
			}
		}
	case reflect.Struct:
		if t.bitmapSize > 0 {
			err = bitmapStructDecode(v, t, val)
			if err != nil {
				return
			}
		} else {
			err = structDecode(v, val)
			if err != nil {
				return
			}
		}
	case reflect.Slice:
		v.SetBytes(val)
	case reflect.String:
		if cp := t.codePage.value(); cp != "" {
			val = decodeUTF8(cp, val)
		}
		v.SetString(string(val))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return nil, err
		}
		v.SetInt(int64(i))
	case reflect.Float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	default:
		return nil, fmt.Errorf("Not support for llvar type")
	}
	return leftBytes, nil
}

func llnumDecoder(v reflect.Value, t tag, data []byte) (leftBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error in llnum")
			return
		}
	}()
	var contentLen int
	switch t.lenEncode {
	case ascii:
		leftBytes = data[2:]
		contentLen, err = strconv.Atoi(string(data[:2]))
		if err != nil {
			err = fmt.Errorf("Parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		leftBytes = data[1:]
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:1], 2)))
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("In llnum, length encoder is invalid")
		return
	}
	val := leftBytes[:contentLen]
	leftBytes = leftBytes[contentLen:]

	switch t.encode {
	case ascii:
	case rbcd:
		fallthrough
	case bcd:
		bcdLen := (contentLen + 1) / 2
		val = bcdl2Ascii(val[:bcdLen], contentLen)
	default:
		err = fmt.Errorf("In llnum, encode not support")
	}

	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(val))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return nil, err
		}
		v.SetInt(int64(i))
	case reflect.Float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	default:
		return nil, fmt.Errorf("Not support for llnum type")
	}

	return leftBytes, nil
}

func lllvarDecoder(v reflect.Value, t tag, data []byte) (leftBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error in lllvar")
			return
		}
	}()
	var contentLen int
	switch t.lenEncode {
	case ascii:
		leftBytes = data[3:]
		contentLen, err = strconv.Atoi(string(data[:3]))
		if err != nil {
			err = fmt.Errorf("Parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		leftBytes = data[2:]
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:2], 3)))
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("In lllvar, length encoder is invalid")
		return
	}
	val := leftBytes[:contentLen]
	leftBytes = leftBytes[contentLen:]
	switch v.Type().Kind() {
	case reflect.Ptr:
		if t.bitmapSize > 0 {
			err = bitmapPtrDecode(v, t, val)
			if err != nil {
				return
			}
		} else {
			err = ptrDecode(v, val)
			if err != nil {
				return
			}
		}
	case reflect.Struct:
		if t.bitmapSize > 0 {
			err = bitmapStructDecode(v, t, val)
			if err != nil {
				return
			}
		} else {
			err = structDecode(v, val)
			if err != nil {
				return
			}
		}
	case reflect.Slice:
		v.SetBytes(val)
	case reflect.String:
		if cp := t.codePage.value(); cp != "" {
			val = decodeUTF8(cp, val)
		}
		v.SetString(string(val))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return nil, err
		}
		v.SetInt(int64(i))
	case reflect.Float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	default:
		return nil, fmt.Errorf("Not support for lllvar type")
	}
	return leftBytes, nil
}

func lllnumDecoder(v reflect.Value, t tag, data []byte) (leftBytes []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error in lllnum")
			return
		}
	}()
	var contentLen int
	switch t.lenEncode {
	case ascii:
		leftBytes = data[3:]
		contentLen, err = strconv.Atoi(string(data[:3]))
		if err != nil {
			err = fmt.Errorf("Parsing length ascii failed: %s", hex.EncodeToString(data))
			return
		}
	case rbcd:
		fallthrough
	case bcd:
		leftBytes = data[2:]
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(data[:2], 3)))
		if err != nil {
			err = fmt.Errorf("Parsing length bcd failed: %s", hex.EncodeToString(data))
			return
		}
	default:
		err = fmt.Errorf("In lllnum, length encoder is invalid")
		return
	}
	val := leftBytes[:contentLen]
	leftBytes = leftBytes[contentLen:]

	switch t.encode {
	case ascii:
	case rbcd:
		fallthrough
	case bcd:
		bcdLen := (contentLen + 1) / 2
		val = bcdl2Ascii(val[:bcdLen], contentLen)
	default:
		err = fmt.Errorf("In lllnum, encode not support")
	}

	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(val))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return nil, err
		}
		v.SetInt(int64(i))
	case reflect.Float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, err
		}
		v.SetFloat(f)
	default:
		return nil, fmt.Errorf("Not support for lllnum type")
	}

	return leftBytes, nil
}

func ptrDecode(v reflect.Value, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			err = fmt.Errorf("critical error, struct decode")
		}
	}()
	t := v.Type().Elem()
	if v.IsNil() {
		v.Set(reflect.New(t))
	}
	var idx int
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		t, err := parseInnerTag(f.Tag)
		if err != nil {
			err = nil
			continue
		}
		err = newValueInnerDecoder(f.Type)(reflect.Indirect(v).Field(i), data[idx:idx+t.length], t.codePage)
		if err != nil {
			break
		}
		idx = idx + t.length
	}
	return
}

func structDecode(v reflect.Value, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("critical error, struct decode")
		}
	}()
	var idx int
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		t, err := parseInnerTag(f.Tag)
		if err != nil {
			err = nil
			continue
		}
		err = newValueInnerDecoder(f.Type)(v.Field(i), data[idx:idx+t.length], t.codePage)
		if err != nil {
			break
		}
		idx = idx + t.length
	}
	return
}

type valueInnerDecoder func(v reflect.Value, data []byte, cp codepageType) error

func newValueInnerDecoder(t reflect.Type) valueInnerDecoder {
	if t == nil {
		return nilInnerDecoder
	}
	switch t.Kind() {
	case reflect.Ptr:
		return ptrDecoder(t)
	case reflect.String:
		return stringInnerDecoder
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intInnerDecoder
	case reflect.Float64:
		return floatInnerDecoder(64)
	case reflect.Float32:
		return floatInnerDecoder(32)
	}
	return unknownInnerDecoder
}

func nilInnerDecoder(v reflect.Value, data []byte, cp codepageType) error {
	return nil
}

func ptrDecoder(t reflect.Type) valueInnerDecoder {
	return func(v reflect.Value, data []byte, cp codepageType) error {
		if len(data) <= 0 {
			return nilInnerDecoder(v, data, cp)
		}
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		return newValueInnerDecoder(v.Elem().Type())(reflect.Indirect(v), data, cp)
	}
}

func unknownInnerDecoder(v reflect.Value, data []byte, cp codepageType) error {
	return fmt.Errorf("Unsupported decoder for field %s", v.Type().Kind().String())
}

func stringInnerDecoder(v reflect.Value, data []byte, cp codepageType) error {
	if c := cp.value(); c != "" {
		data = decodeUTF8(c, data)
	}
	data = bytes.TrimSpace(data)
	v.SetString(string(data))
	return nil
}

func intInnerDecoder(v reflect.Value, data []byte, cp codepageType) error {
	if len(data) < 1 {
		return nil
	}
	i, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func floatInnerDecoder(bitSize int) valueInnerDecoder {
	return func(v reflect.Value, data []byte, cp codepageType) error {
		if len(data) < 1 {
			return nil
		}
		f, err := strconv.ParseFloat(string(data), bitSize)
		if err != nil {
			return err
		}
		v.SetFloat(f)
		return nil
	}
}
