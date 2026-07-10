package idconv

import (
	"testing"

	"github.com/google/uuid"
)

func TestFromUUID(t *testing.T) {
	cases := map[string]uint{
		"00000000-0000-0000-0000-000000000000": 0,
		"00000000-0000-0000-0000-000000000001": 1,
		"00000000-0000-0000-0000-0000000000ff": 255,
	}
	for s, want := range cases {
		if got := FromUUID(uuid.MustParse(s)); got != want {
			t.Errorf("FromUUID(%s) = %d, attendu %d", s, got, want)
		}
	}
}

func TestFromUUIDNilReturnsZero(t *testing.T) {
	if got := FromUUID(uuid.Nil); got != 0 {
		t.Errorf("FromUUID(Nil) = %d, attendu 0", got)
	}
}

func TestToUUIDRoundTrip(t *testing.T) {
	for _, id := range []uint{0, 1, 255, 4096} {
		if got := FromUUID(ToUUID(id)); got != id {
			t.Errorf("FromUUID(ToUUID(%d)) = %d, attendu %d", id, got, id)
		}
	}
}

func TestToUUID(t *testing.T) {
	cases := map[uint]string{
		0:   "00000000-0000-0000-0000-000000000000",
		1:   "00000000-0000-0000-0000-000000000001",
		255: "00000000-0000-0000-0000-0000000000ff",
	}
	for id, want := range cases {
		if got := ToUUID(id).String(); got != want {
			t.Errorf("ToUUID(%d) = %s, attendu %s", id, got, want)
		}
	}
}
