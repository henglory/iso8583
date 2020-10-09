package iso8583v2

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

func Unmarshal(data []byte, v interface{}) error {
	rv, err := validateDecode(v)
	if err != nil {
		return fmt.Errorf("validate failed %s", err.Error())
	}

	//lookup for tag in cache
	structKey := rv.Type().PkgPath() + rv.Type().Name()
	mpTag := tagCache[structKey]
	tag, ok := mpTag.Load().(map[string]*iso8583Tag)
	if !ok {
		var tagValue atomic.Value
		tag = loadTag(rv)
		tagValue.Store(tag)
		tagCache[structKey] = tagValue
	}
	return decodeIso8583wthTag(data, rv, tag)
}

func decodeIso8583wthTag(data []byte, v reflect.Value, tag map[string]*iso8583Tag) error {
	d := &decoder{}
	for i := 0; i < v.Type().NumField(); i++ {
		field := v.Type().Field(i)
		isoTag := tag[field.Name]
		if isoTag == nil {
			continue
		}
		if isoTag.isMti {
			d.setMtiDecoder(&mtiDecoder{
				v:  v.Field(i),
				tg: *isoTag,
			})
			d.setBitmapDecoder(&bitmapDecoder{})
			continue
		}
		d.addFieldDecoder(&fieldDecoder{
			v:  v.Field(i),
			tg: *isoTag,
		})

	}
	return d.execute(data)
}

func validateDecode(v interface{}) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return rv, fmt.Errorf("Unmarshaling struct must be a pointer")
	}
	for rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}
	if rv.Kind() == reflect.Slice {
		return rv, fmt.Errorf("Decode unsupport slice")
	}
	return rv, nil
}
