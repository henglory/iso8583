package iso8583

import (
	"encoding/hex"
	"testing"
)

func TestCalculateBitmap(t *testing.T) {
	var bitmap = make([]byte, 8)
	bitmap, err := calculateBitmap(bitmap, 1)
	if err != nil {
		t.Errorf("Should not be error %+v", err)
	}
	if hex.EncodeToString(bitmap) != "8000000000000000" {
		t.Errorf("bitmap calculate fail %s", hex.EncodeToString(bitmap))
	}
	bitmap, err = calculateBitmap(bitmap, 4)
	if hex.EncodeToString(bitmap) != "9000000000000000" {
		t.Errorf("bitmap calculate fail %s", hex.EncodeToString(bitmap))
	}
	bitmap, err = calculateBitmap(bitmap, 64)
	if hex.EncodeToString(bitmap) != "9000000000000001" {
		t.Errorf("bitmap calculate fail %s", hex.EncodeToString(bitmap))
	}
}
