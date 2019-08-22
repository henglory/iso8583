package iso8583

import (
	"reflect"
	"testing"
)

type fakeTagStruct struct {
	Mti string `encode:"bcd"`
	T1  string `field:"4" length:"10" encode:"ascii" type:"llvar"`
}

func TestParseTag(t *testing.T) {
	s := fakeTagStruct{
		Mti: "0800",
	}
	v := reflect.ValueOf(s)
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		tg, err := parseTag(f)
		if err != nil {
			t.Errorf("parse tag fail %s", err.Error())
		}
		if f.Name == "T1" && (tg.field != 4 || tg.length != 10 || tg.encode != ascii || tg.fieldType != llvar) {
			t.Errorf("parse tag fail %+v", tg)
		}
		if f.Name == "Mti" && (tg.encode != bcd || tg.isMti != true) {
			t.Errorf("parse tag fail %+v", tg)
		}
	}

}
