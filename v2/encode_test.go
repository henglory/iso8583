package iso8583v2

import (
	"encoding/hex"
	"testing"
)

type test1 struct {
	Test  string
	Test2 string
}

type test2 struct {
	Mti string `encode:"bcd"`
	T1  string `field:"4" encode:"ascii" type:"llvar"`
	T2  string `field:"5" encode:"lbcd," type:"lllvar"`
	T3  string `field:"6" length:"10" encode:"ascii" type:"numeric"`
}

func TestMarshal(t *testing.T) {
	s1 := test2{
		Mti: "0200",
		T1:  "123123",
		T2:  "123",
	}
	_, err := Marshal(s1)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateEncode(t *testing.T) {
	s1 := struct {
		Test  string
		Test2 string
	}{
		Test:  "test",
		Test2: "test2",
	}
	_, err := validateEncode(s1)
	if err != nil {
		t.Error("normal struct validate not passed", err)
	}
	_, err = validateEncode(&s1)
	if err != nil {
		t.Error("normal struct with pointer validate not passed", err)
	}
	var f1 *struct{}
	_, err = validateEncode(f1)
	if err == nil {
		t.Error("nil pointer validate pass")
	}
	var f2 []struct{}
	_, err = validateEncode(f2)
	if err == nil {
		t.Error("slice should not be supported")
	}

	var i int
	_, err = validateEncode(i)
	if err == nil {
		t.Error("support only struct")
	}
}

type T48 struct {
	T1 string `field:"1" length:"10" cp:"TIS-620"`
	T2 string `field:"2" length:"5"`
	T3 int    `field:"3" length:"9"`
}

type TestIso struct {
	Mti         string `encode:"bcd"`
	TransmissDt string `field:"7" length:"10" type:"numeric"`
	TraceNum    string `field:"11" length:"6" type:"numeric"`
	SendingID   string `field:"32" type:"llvar"`
	Rrn         string `field:"37" length:"12" type:"numeric"`
	T           T48    `field:"48" type:"llvar" encode:"bcd,ascii"`
	NetworkCode string `field:"70" length:"3" type:"numeric"`
}

type TestBitmapIso struct {
	Mti         string `encode:"bcd"`
	TransmissDt string `field:"7" length:"10" type:"numeric"`
	TraceNum    string `field:"11" length:"6" type:"numeric"`
	SendingID   string `field:"32" type:"llvar"`
	Rrn         string `field:"37" length:"12" type:"numeric"`
	T           T48    `field:"48" type:"llvar" encode:"bcd,ascii" bitmapsize:"8"`
	NetworkCode string `field:"70" length:"3" type:"numeric"`
}

func TestEncode(t *testing.T) {

	init := TestIso{
		Mti:         "0800",
		TransmissDt: "123123",
		TraceNum:    "123456",
		SendingID:   "004",
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
	if hex.EncodeToString(b) != "08008220000108010000040000000000000030303030313233313233313233343536303330303430303039303832333231323324b7b4cacdbab7b4cacdba3132332020303030303030303031303830" {
		t.Errorf("Marshal data error %s\n", hex.EncodeToString(b))
	}
}

func TestBitmapEncode(t *testing.T) {

	init := TestBitmapIso{
		Mti:         "0800",
		TransmissDt: "123123",
		TraceNum:    "123456",
		SendingID:   "004",
		T: T48{
			T1: "ทดสอบทดสอบ",
			T2: "",
			T3: 1,
		},
		Rrn:         "908232123",
		NetworkCode: "80",
	}
	b, err := Marshal(init)
	if err != nil {
		t.Errorf("marshal error %+v", err)
	}
	if hex.EncodeToString(b) != "08008220000108010000040000000000000030303030313233313233313233343536303330303430303039303832333231323327a000000000000000b7b4cacdbab7b4cacdba303030303030303031303830" {
		t.Errorf("encode with bitmap failed %s", hex.EncodeToString(b))
	}

}
