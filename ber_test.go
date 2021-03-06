package pkcs7

import (
	"bytes"
	"encoding/asn1"
	"strings"
	"testing"
)

func TestBer2Der(t *testing.T) {
	// indefinite length fixture
	ber := []byte{0x30, 0x80, 0x02, 0x01, 0x01, 0x00, 0x00}
	expected := []byte{0x30, 0x03, 0x02, 0x01, 0x01}
	der, err := ber2der(ber)
	if err != nil {
		t.Fatalf("ber2der failed with error: %v", err)
	}
	if bytes.Compare(der, expected) != 0 {
		t.Errorf("ber2der result did not match.\n\tExpected: % X\n\tActual: % X", expected, der)
	}

	if der2, err := ber2der(der); err != nil {
		t.Errorf("ber2der on DER bytes failed with error: %v", err)
	} else {
		if !bytes.Equal(der, der2) {
			t.Error("ber2der is not idempotent")
		}
	}
	var thing struct {
		Number int
	}
	rest, err := asn1.Unmarshal(der, &thing)
	if err != nil {
		t.Errorf("Cannot parse resulting DER because: %v", err)
	} else if len(rest) > 0 {
		t.Errorf("Resulting DER has trailing data: % X", rest)
	}
}

func TestBer2Der_Negatives(t *testing.T) {
	fixtures := []struct {
		Input         []byte
		ErrorContains string
	}{
		{[]byte{0x30, 0x85}, "length too long"},
		{[]byte{0x30, 0x84, 0x80, 0x0, 0x0, 0x0}, "length is negative"},
		{[]byte{0x30, 0x82, 0x0, 0x1}, "length has leading zero"},
		{[]byte{0x30, 0x80, 0x1, 0x2, 0x1, 0x2}, "Invalid BER format"},
		{[]byte{0x30, 0x03, 0x01, 0x02}, "length is more than available data"},
	}

	for _, fixture := range fixtures {
		_, err := ber2der(fixture.Input)
		if err == nil {
			t.Errorf("No error thrown. Expected: %s", fixture.ErrorContains)
		}
		if !strings.Contains(err.Error(), fixture.ErrorContains) {
			t.Errorf("Unexpected error thrown.\n\tExpected: /%s/\n\tActual: %s", fixture.ErrorContains, err.Error())
		}
	}
}

func TestBer2Der_NestedMultipleIndefinite(t *testing.T) {
	// indefinite length fixture
	ber := []byte{0x30, 0x80, 0x30, 0x80, 0x02, 0x01, 0x01, 0x00, 0x00, 0x30, 0x80, 0x02, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00}
	expected := []byte{0x30, 0x0A, 0x30, 0x03, 0x02, 0x01, 0x01, 0x30, 0x03, 0x02, 0x01, 0x02}

	der, err := ber2der(ber)
	if err != nil {
		t.Fatalf("ber2der failed with error: %v", err)
	}
	if bytes.Compare(der, expected) != 0 {
		t.Errorf("ber2der result did not match.\n\tExpected: % X\n\tActual: % X", expected, der)
	}

	if der2, err := ber2der(der); err != nil {
		t.Errorf("ber2der on DER bytes failed with error: %v", err)
	} else {
		if !bytes.Equal(der, der2) {
			t.Error("ber2der is not idempotent")
		}
	}
	var thing struct {
		Nest1 struct {
			Number int
		}
		Nest2 struct {
			Number int
		}
	}
	rest, err := asn1.Unmarshal(der, &thing)
	if err != nil {
		t.Errorf("Cannot parse resulting DER because: %v", err)
	} else if len(rest) > 0 {
		t.Errorf("Resulting DER has trailing data: % X", rest)
	}
}

func TestBer2Der_ConstructedStrings(t *testing.T) {
	var ber_arr [][]byte
	var octedExpected_arr [][]byte
	ber_arr = [][]byte{

		{0x23, 0x09, 0x03, 0x03, 0x00, 0x6e, //bit string
			0x5d, 0x03, 0x02, 0x06, 0xc0},
		{0x24, 0x0c, 0x04, 0x04, 0x01, 0x23, 0x45, //octet string
			0x67, 0x04, 0x04, 0x89, 0xab, 0xcd, 0xef},
		{0x33, 0x0f, 0x13, 0x05, 0x54, //PrintableString
			0x65, 0x73, 0x74, 0x20, 0x13,
			0x06, 0x55, 0x73, 0x65, 0x72,
			0x20, 0x31},
		{0x34, 0x15, 0x14, 0x05, 0x63, // T61String
			0x6c, 0xc2, 0x65, 0x73, 0x14,
			0x01, 0x20, 0x14, 0x09, 0x70,
			0x75, 0x62, 0x6c, 0x69, 0x71,
			0x75, 0x65, 0x73},
		{0x36, 0x13, 0x16, 0x05, 0x74, //IA5String
			0x65, 0x73, 0x74, 0x31, 0x16,
			0x01, 0x40, 0x16, 0x07, 0x72,
			0x73, 0x61, 0x2e, 0x63, 0x6f,
			0x6d},
	}

	octedExpected_arr = [][]byte{
		{0x23, 0x07, 0x03, 0x05, 0x00, 0x6e, //bit string
			0x5d, 0x06, 0xc0},
		{0x24, 0xa, 0x04, 0x08, 0x01, 0x23, //octed string
			0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x33, 0x0d, 0x13, 0x0b, 0x54, //PrintableString
			0x65, 0x73, 0x74, 0x20, 0x55,
			0x73, 0x65, 0x72, 0x20, 0x31},
		{0x34, 0x11, 0x14, 0x0F, 0x63, // T61String
			0x6c, 0xc2, 0x65, 0x73, 0x20,
			0x70, 0x75, 0x62, 0x6c, 0x69,
			0x71, 0x75, 0x65, 0x73},
		{0x36, 0x0F, 0x16, 0x0d, 0x74, //IA5String
			0x65, 0x73, 0x74, 0x31, 0x40,
			0x72, 0x73, 0x61, 0x2e, 0x63,
			0x6f, 0x6d},
	}

	for idx, ber := range ber_arr {
		der, err := ber2der(ber)
		if err != nil {
			t.Fatalf("ber2der failed with error: %v", err)
		}
		if bytes.Compare(der, octedExpected_arr[idx]) != 0 {
			t.Errorf("ber2der result did not match.\n\tExpected: % X\n\tActual: % X",
				octedExpected_arr[idx], der)
		}
	}
}
