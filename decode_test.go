package iso8583

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	init := TestIso{
		Mti:          "0800",
		SecondBitmap: true,
		TransmissDt:  "123123",
		TraceNum:     "123456",
		SendingID:    "004",
		T: T48{
			T1: "ทดสอบทดสอบ",
			T2: "123",
			T3: 1,
		},
		Rrn:         "908232123",
		NetworkCode: "80",
	}
	b, err := Marshal(init)
	if err != nil {
		t.Errorf("marshal error %+v", err)
	}
	iso := TestIso{}
	//need to cut first 11 byte for dataLen 2 bytes & ISO header 9 bytes (ISO00(len2))
	err = Unmarshal(b, &iso)
	if err != nil {
		t.Errorf("unmarshal error %+v", err)
	}
	if !reflect.DeepEqual(init, iso) {
		t.Errorf("marshal & unmarshal error %+v\n %+v\n", init, iso)
	}
}
