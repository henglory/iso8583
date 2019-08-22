package iso8583

import (
	"encoding/hex"
)

func lbcdEncode(data []byte) []byte {
	if len(data)%2 != 0 {
		return bcdEncode(append(data, "0"...))
	}
	return bcdEncode(data)
}

func rbcdEncode(data []byte) []byte {
	if len(data)%2 != 0 {
		return bcdEncode(append([]byte("0"), data...))
	}
	return bcdEncode(data)
}

func bcdEncode(data []byte) []byte {
	out := make([]byte, len(data)/2+1)
	n, err := hex.Decode(out, data)
	if err != nil {
		panic(err.Error())
	}
	return out[:n]
}

func bcdl2Ascii(data []byte, length int) []byte {
	return bcd2Ascii(data)[:length]
}

func bcdr2Ascii(data []byte, length int) []byte {
	out := bcd2Ascii(data)
	return out[len(out)-length:]
}

func bcd2Ascii(data []byte) []byte {
	out := make([]byte, len(data)*2)
	n := hex.Encode(out, data)
	return out[:n]
}
