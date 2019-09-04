package iso8583

import (
	"reflect"
	"testing"
)

type TestIsoDecode struct {
	Mti          string `encode:"bcd"`
	SecondBitmap bool
	TransmissDt  string `field:"7" length:"10" type:"numeric"`
	TraceNum     string `field:"11" length:"6" type:"numeric"`
	SendingID    string `field:"32" type:"llvar"`
	Rrn          string `field:"37" length:"12" type:"numeric"`
	T            T48    `field:"48" type:"lllvar" encode:"bcd,ascii"`
	NetworkCode  string `field:"70" length:"3" type:"numeric"`
}

func TestDecode(t *testing.T) {
	init := TestIsoDecode{
		Mti:          "0800",
		SecondBitmap: true,
		TransmissDt:  "0000123123",
		TraceNum:     "123456",
		SendingID:    "004",
		T: T48{
			T1: "ทดสอบทดสอบ",
			T2: "123",
			T3: 1,
		},
		Rrn:         "000908232123",
		NetworkCode: "080",
	}
	b, err := Marshal(init)
	if err != nil {
		t.Errorf("marshal error %+v", err)
	}
	iso := TestIsoDecode{}

	err = Unmarshal(b, &iso)
	if err != nil {
		t.Errorf("unmarshal error %+v", err)
	}
	if !reflect.DeepEqual(init, iso) {
		t.Errorf("marshal & unmarshal error %+v\n %+v\n", init, iso)
	}
}
