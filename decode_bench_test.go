package iso8583

import (
	"encoding/hex"
	"testing"
)

func BenchmarkUnmarshal_1000000(b *testing.B) {
	data, _ := hex.DecodeString("0800022000010801000030303030313233313233313233343536303330303430303039303832333231323324b7b4cacdbab7b4cacdba3132332020303030303030303031")
	var t TestIso
	b.ResetTimer()
	for j := 0; j < b.N; j++ {

		Unmarshal(data, &t)
	}
}
