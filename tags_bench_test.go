package iso8583

import (
	"reflect"
	"testing"
)

func BenchmarkTags(b *testing.B) {

	s := TestIso{
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
	v := reflect.ValueOf(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < v.Type().NumField(); i++ {
			f := v.Type().Field(i)
			parseTag(f)
		}
	}
}
