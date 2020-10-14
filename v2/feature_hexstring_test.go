package v2

import (
	"fmt"
	"testing"
)

type hexTestStruct struct {
	Mti string
	Emv string `field:"2" type:"lllvar" cp:"hexstring"`
}

func TestCompareBeforeAndAfter(t *testing.T) {

	before := hexTestStruct{
		Mti: "0200",
		Emv: "ea4709d9f36020",
	}

	b, err := Marshal(before)
	if err != nil {
		t.Error("fail marshal test data", err.Error())
	}

	after := hexTestStruct{}

	err = Unmarshal(b, &after)
	if err != nil {
		t.Error("fail unmarshal test data", err.Error())
	}

	if fmt.Sprintf("%#v", before) != fmt.Sprintf("%#v", after) {
		t.Error("fail hexstring encode/decode is malfunction")
	}

}
