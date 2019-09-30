package iso8583

import "testing"

func TestCheckIsOn(t *testing.T) {
	var bitmap = make([]byte, 8)
	bitmap, err := calculateBitmap(bitmap, 1)
	if err != nil {
		t.Error(err)
	}
	isOn, err := bitIsOn(bitmap, 1)
	if err != nil {
		t.Error(err)
	}
	if !isOn {
		t.Errorf("bit 1 should be on")
	}
	isOn, err = bitIsOn(bitmap, 2)
	if err != nil {
		t.Error(err)
	}
	if isOn {
		t.Errorf("bit 2 should be not on")
	}
}
