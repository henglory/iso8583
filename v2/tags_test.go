package iso8583v2

import (
	"reflect"
	"testing"
)

type fakeTagStruct struct {
	Mti string `encode:"bcd"`
	T1  string `field:"4" encode:"ascii" type:"llvar"`
	T2  string `field:"5" encode:"lbcd,lbcd" type:"lllvar"`
	T3  string `field:"6" length:"10" encode:"ascii" type:"numeric"`
}

func TestParseTag(t *testing.T) {
	s := fakeTagStruct{
		Mti: "0800",
		T1:  "123123",
	}
	v := reflect.ValueOf(s)
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		tg, err := parseIso8583Tag(f)
		if err != nil {
			t.Errorf("parse tag fail %s", err.Error())
		}
		if f.Name == "T1" && (tg.field != 4 || tg.valEncode != ascii || tg.fieldType != llvar) {
			t.Errorf("parse tag fail %+v", tg)
		}
		if f.Name == "T2" && (tg.field != 5 || tg.valEncode != bcd || tg.fieldType != lllvar) {
			t.Errorf("parse tag fail %+v", tg)
		}
		if f.Name == "T3" && (tg.field != 6 || tg.valEncode != ascii || tg.length != 10) {
			t.Errorf("parse tag fail %+v", tg)
		}
		if f.Name == "Mti" && (tg.valEncode != bcd || tg.isMti != true) {
			t.Errorf("parse tag fail %+v", tg)
		}
	}

}
