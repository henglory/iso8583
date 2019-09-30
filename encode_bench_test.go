package iso8583

import (
	"reflect"
	"testing"
)

func BenchmarkMarshal(b *testing.B) {

	init := TestIso{
		Mti:          "0800",
		SecondBitmap: false,
		TransmissDt:  "123123",
		TraceNum:     "123456",
		SendingID:    "004",
		T: T48{
			T1: "ทดสอบทดสอบจ้า",
			T2: "123",
			T3: 1,
		},
		Rrn:         "908232123",
		NetworkCode: "80",
	}

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		_, err := Marshal(init)
		if err != nil {
			b.Errorf("marshal on benchmark error %+v", err)
		}
	}
}

func BenchmarkInitEncode(b *testing.B) {

	init := TestIso{
		Mti:          "0800",
		SecondBitmap: false,
		TransmissDt:  "123123",
		TraceNum:     "123456",
		SendingID:    "004",
		T: T48{
			T1: "ทดสอบทดสอบจ้า",
			T2: "123",
			T3: 1,
		},
		Rrn:         "908232123",
		NetworkCode: "80",
	}
	v := reflect.ValueOf(init)
	for i := 0; i < b.N; i++ {
		initEncoder(v)
	}
}
