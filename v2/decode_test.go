package iso8583v2

import (
	"encoding/hex"
	"reflect"
	"testing"
)

type TestIsoDecode struct {
	Mti         string `encode:"bcd"`
	TransmissDt string `field:"7" length:"10" type:"numeric"`
	TraceNum    string `field:"11" length:"6" type:"numeric"`
	SendingID   string `field:"32" type:"llvar"`
	Rrn         string `field:"37" length:"12" type:"numeric"`
	T           T48    `field:"48" type:"lllvar" encode:"bcd,ascii"`
	NetworkCode string `field:"70" length:"3" type:"numeric"`
}

type TestBitmapIsoDecode struct {
	Mti         string `encode:"bcd"`
	TransmissDt string `field:"7" length:"10" type:"numeric"`
	TraceNum    string `field:"11" length:"6" type:"numeric"`
	SendingID   string `field:"32" type:"llvar"`
	Rrn         string `field:"37" length:"12" type:"numeric"`
	T           T48    `field:"48" type:"llvar" encode:"bcd,ascii" bitmapsize:"8"`
	NetworkCode string `field:"70" length:"3" type:"numeric"`
}

type testIsoDecodeSubFieldPointerStruct struct {
	Mti         string         `encode:"bcd"`
	TransmissDt string         `field:"7" length:"10" type:"numeric"`
	TraceNum    string         `field:"11" length:"6" type:"numeric"`
	SendingID   string         `field:"32" type:"llvar"`
	Rrn         string         `field:"37" length:"12" type:"numeric"`
	T           t48WithPointer `field:"48" type:"llvar" encode:"bcd,ascii" bitmapsize:"8"`
	NetworkCode *string        `field:"70" length:"3" type:"numeric"`
}

type t48WithPointer struct {
	T1 string  `field:"1" length:"10" cp:"TIS-620"`
	T2 *string `field:"2" length:"5"`
	T3 int     `field:"3" length:"9"`
}

func TestIsoDecodeSubFieldPointer(t *testing.T) {
	init := testIsoDecodeSubFieldPointerStruct{
		Mti:         "0200",
		TransmissDt: "2020010112",
		TraceNum:    "123345",
		SendingID:   "004",
		Rrn:         "123456789012",
		T: t48WithPointer{
			T1: "นายทดสอบdd",
			T3: 10,
		},
	}
	nw := "080"
	init.NetworkCode = &nw
	init.T.T2 = &nw
	b, err := Marshal(init)
	if err != nil {
		t.Error(err)
	}
	if hex.EncodeToString(b) != "02008220000108010000040000000000000032303230303130313132313233333435303330303431323334353637383930313232e000000000000000b9d2c2b7b4cacdba64643038302020303030303030303130303830" {
		t.Error("marshaling incorrect")
	}
	bh, _ := hex.DecodeString("02008220000108010000040000000000000032303230303130313132313233333435303330303431323334353637383930313232e000000000000000b9d2c2b7b4cacdba64643038302020303030303030303130303830")
	iso := testIsoDecodeSubFieldPointerStruct{}
	err = Unmarshal(bh, &iso)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(init, iso) {
		t.Errorf("marshal & unmarshal error %+v\n %+v\n", init, iso)
	}
}

//02008220000108010000040000000000000032303230303130313132313233333435303330303431323334353637383930313232e000000000000000b9d2c2b7b4cacdba64643038302020303030303030303130303830
//02008220000108010000040000000000000032303230303130313132313233333435303330303431323334353637383930313227a000000000000000b9d2c2b7b4cacdba6464303030303030303130303830
//0200
//82200001080100000400000000000000 1000 0010 0010 0000 0000 0000 0000 0001 0000 1000
//32303230303130313132
//313233333435
//3033 303034
//313233343536373839303132
//27
//1010
//a000000000000000b9d2c2b7b4cacdba6464303030303030303130
//303830

func TestDecodeWithPointer(t *testing.T) {
	init := test3{
		Mti: "0200",
		T1:  "test",
		T3:  "0000000123",
	}
	s := "test2"
	init.T2 = &s
	b, _ := hex.DecodeString("303230301c000000000000003034746573740005746573743230303030303030313233")
	iso := test3{}
	err := Unmarshal(b, &iso)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(init, iso) {
		t.Errorf("marshal & unmarshal error %+v\n %+v\n", init, iso)
	}
}

func TestDecode(t *testing.T) {

	init := TestIsoDecode{
		Mti:         "0800",
		TransmissDt: "0000123123",
		TraceNum:    "123456",
		SendingID:   "004",
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

func TestBitmapDecode(t *testing.T) {

	init := TestBitmapIsoDecode{
		Mti:         "0800",
		TransmissDt: "0000123123",
		TraceNum:    "123456",
		SendingID:   "004",
		T: T48{
			T1: "ทดสอบทดสอบ",
			T2: "",
			T3: 1,
		},
		Rrn:         "000908232123",
		NetworkCode: "080",
	}
	b, err := Marshal(init)
	if err != nil {
		t.Errorf("marshal error %+v", err)
	}
	if hex.EncodeToString(b) != "08008220000108010000040000000000000030303030313233313233313233343536303330303430303039303832333231323327a000000000000000b7b4cacdbab7b4cacdba303030303030303031303830" {
		t.Errorf("encode with bitmap failed %s", hex.EncodeToString(b))
	}

	iso := TestBitmapIsoDecode{}
	err = Unmarshal(b, &iso)
	if err != nil {
		t.Errorf("unmarshal error %+v", err)
	}
	if !reflect.DeepEqual(init, iso) {
		t.Errorf("should be equal \n%+v\n%+v", init, iso)
	}
}
