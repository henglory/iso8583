package iso8583v2

import (
	"fmt"
	"reflect"
)

type fieldDecoder struct {
	getBitmap func() []byte
	v         reflect.Value
	tg        iso8583Tag
}

func (f *fieldDecoder) decode(data []byte) ([]byte, error) {
	bitmap := f.getBitmap()
	if bitmap == nil {
		return nil, fmt.Errorf("decode failed field:%s bitmap data not found", f.tg.name)
	}
	maxField := len(bitmap) * 8
	if f.tg.field > maxField || len(data) == 0 {
		return data, nil
	}
	byteIndex := (f.tg.field - 1) / 8
	bitIndex := (f.tg.field - 1) % 8
	if (bitmap[byteIndex] & (0x80 >> uint(bitIndex))) != (0x80 >> uint(bitIndex)) {
		//offbit
		return data, nil
	}
	//onbit
	switch f.tg.fieldType {
	case numeric:
		return f.numericDecode(data)
	case alpha:
		return f.alphaDecode(data)
	case binary:
		return f.binaryDecode(data)
	case llvar:
		return f.llvarDecode(data)
	case lllvar:
		return f.lllvarDecode(data)
	default:
		return nil, fmt.Errorf("decode failed field:%s unknown field type %s", f.tg.name, f.tg.fieldType.value())
	}
}

func (f *fieldDecoder) getValueEncoderFn() func(data []byte) ([]byte, []byte, error) {
	switch f.tg.valEncode {
	case bcd:
		return f.bcdDecoder
	case rbcd:
		return f.rbcdDecode
	case ascii:
		return f.asciiDecode
	case hexstring:
		return f.hexstringDecode
	default:
		return f.unknownDecode
	}
}

func (f *fieldDecoder) bcdDecoder(data []byte) ([]byte, []byte, error) {
	l := (f.tg.length + 1) / 2
	if len(data) < l {
		return nil, nil, fmt.Errorf("field:%s bcd decode length data is smaller than expected", f.tg.name)
	}
	return bcdl2Ascii(data[:l], f.tg.length), data[l:], nil
}

func (f *fieldDecoder) rbcdDecode(data []byte) ([]byte, []byte, error) {
	l := (f.tg.length + 1) / 2
	if len(data) < l {
		return nil, nil, fmt.Errorf("field:%s rbcd decode length data is smaller than expected", f.tg.name)
	}
	return bcdr2Ascii(data[:l], f.tg.length), data[l:], nil
}

func (f *fieldDecoder) asciiDecode(data []byte) ([]byte, []byte, error) {
	if len(data) < f.tg.length {
		return nil, nil, fmt.Errorf("field:%s ascii decode length data is smaller than expected", f.tg.name)
	}
	return data[:f.tg.length], data[f.tg.length:], nil
}

func (f *fieldDecoder) hexstringDecode(data []byte) ([]byte, []byte, error) {
	if len(data) < f.tg.length {
		return nil, nil, fmt.Errorf("field:%s hexstring decode length data is smaller than expected", f.tg.name)
	}
	return data[:f.tg.length], data[f.tg.length:], nil
}

func (f *fieldDecoder) unknownDecode(data []byte) ([]byte, []byte, error) {
	return nil, nil, fmt.Errorf("decode failed field:%s value encode unsupported %s", f.tg.name, f.tg.valEncode.value())
}

type bitmapDecoder struct {
	bitmap []byte
}

func (b *bitmapDecoder) decode(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("decode failed: bitmap data length too small (%d)", len(data))
	}
	byteNum := 8
	if data[0]&0x80 == 0x80 {
		byteNum = 16
	}
	b.bitmap = data[:byteNum]
	return data[byteNum:], nil
}

func (b *bitmapDecoder) getBitmap() []byte {
	return b.bitmap
}

type mtiDecoder struct {
	v  reflect.Value
	tg iso8583Tag
}

func (m *mtiDecoder) decode(data []byte) ([]byte, error) {
	if m.v.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("mti should be string %v", m.v.Type().Kind())
	}
	switch m.tg.valEncode {
	case ascii:
		if len(data) < 4 {
			return nil, fmt.Errorf("decode ascii failed: mti data length too small (%d)", len(data))
		}
		m.v.SetString(string(data[:4]))
		return data[4:], nil
	case bcd:
		if len(data) < 2 {
			return nil, fmt.Errorf("decode bcd failed: mti data length too small (%d)", len(data))
		}
		m.v.SetString(string(bcd2Ascii(data[:2])))
		return data[2:], nil
	case hexstring:
		if len(data) < 4 {
			return nil, fmt.Errorf("decode hexstring failed: mti data length too small (%d)", len(data))
		}
		m.v.SetString(string(data[:4]))
		return data[4:], nil
	}
	return nil, fmt.Errorf("decode failed: mti field encode value not supported, %s", m.tg.valEncode.value())
}

type decoder struct {
	m   *mtiDecoder
	b   *bitmapDecoder
	fds []*fieldDecoder
}

func (d *decoder) setMtiDecoder(m *mtiDecoder) {
	d.m = m
}

func (d *decoder) setBitmapDecoder(b *bitmapDecoder) {
	d.b = b
}

func (d *decoder) getBitmap() []byte {
	if d.b != nil {
		return d.b.getBitmap()
	}
	return nil
}

func (d *decoder) addFieldDecoder(f *fieldDecoder) {
	f.getBitmap = d.getBitmap
	d.fds = append(d.fds, f)
}

func (d *decoder) execute(data []byte) error {
	var err error
	if d.m == nil {
		return fmt.Errorf("mti decoder is nil")
	}
	data, err = d.m.decode(data)
	if err != nil {
		return fmt.Errorf("decode mti failed %s", err.Error())
	}
	data, err = d.b.decode(data)
	if err != nil {
		return fmt.Errorf("decode bitmap failed %s", err.Error())
	}
	for _, fd := range d.fds {
		if fd != nil {
			data, err = fd.decode(data)
			if err != nil {
				return fmt.Errorf("decode field:%s failed %s", fd.tg.name, err.Error())
			}
		}
	}
	return nil
}
