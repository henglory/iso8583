package iso8583

import (
	"fmt"
	"reflect"
)

type decoder struct {
	data []byte
}

//Unmarshal is used for convert byte array to iso8583 struct
func Unmarshal(data []byte, v interface{}) error {
	return newDecoder(data).decode(v)
}

func newDecoder(b []byte) *decoder {
	return &decoder{
		data: b,
	}
}

func (d *decoder) decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("Decode struct should be a pointer")
	}
	if reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Slice {
		return fmt.Errorf("Decode unsupport slice")
	}
	err := d.read(reflect.Indirect(rv))
	if err != nil {
		return err
	}
	return nil
}

func (d *decoder) read(v reflect.Value) error {
	if reflect.Value(v).Kind() == reflect.Ptr {
		return d.read(reflect.Indirect(v))
	}
	err := initDecoder(v, d.data)
	if err != nil {
		return err
	}
	return nil
}
