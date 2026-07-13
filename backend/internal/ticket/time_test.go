package ticket

import (
	"testing"
	"time"
)

func TestMakassarNow_ReturnsTimeInMakassar(t *testing.T) {
	_, err := time.LoadLocation(MakassarTimezone)
	if err != nil {
		t.Skipf("timezone %s not available: %v", MakassarTimezone, err)
	}

	now := MakassarNow()
	zoneName, offset := now.Zone()

	// UTC+8 = 28800 seconds
	if offset != 8*3600 {
		t.Errorf("expected offset %d seconds (+08:00), got %d", 8*3600, offset)
	}

	// The zone name should be WITA (Western Indonesia Time)
	if zoneName != "WITA" {
		t.Logf("zone name is %q (expected WITA, but this is locale-dependent)", zoneName)
	}

	// Verify the time is within a reasonable range of now
	utcNow := time.Now().UTC()
	diff := utcNow.Sub(now.UTC())
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("MakassarNow() should be close to UTC now, diff=%v", diff)
	}
}

func TestMakassarNow_ConsistentCalls(t *testing.T) {
	// Two consecutive calls should return the same timezone
	t1 := MakassarNow()
	t2 := MakassarNow()

	_, off1 := t1.Zone()
	_, off2 := t2.Zone()

	if off1 != off2 {
		t.Errorf("timezone offset changed between calls: %d vs %d", off1, off2)
	}
}

func TestTodayMakassarBoundary_ReturnsStartOfDay(t *testing.T) {
	loc, err := time.LoadLocation(MakassarTimezone)
	if err != nil {
		t.Skipf("timezone %s not available: %v", MakassarTimezone, err)
	}

	boundary := TodayMakassarBoundary()

	// Should be midnight in Makassar
	h, m, s := boundary.Clock()
	if h != 0 || m != 0 || s != 0 {
		t.Errorf("expected 00:00:00, got %02d:%02d:%02d", h, m, s)
	}

	// Should be in the Makassar timezone
	zoneName, offset := boundary.Zone()
	_ = zoneName
	if offset != 8*3600 {
		t.Errorf("expected offset +08:00, got %d", offset)
	}

	// The boundary should be today's date in Makassar
	now := time.Now().In(loc)
	expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if !boundary.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, boundary)
	}
}

func TestTodayMakassarBoundary_BeforeAndAfterMidnight(t *testing.T) {
	// This test verifies the concept: the boundary should be the start
	// of the current day in Makassar time, regardless of when it's called.
	// We can't easily mock time, but we can verify the structure.
	boundary := TodayMakassarBoundary()

	// Nanoseconds should be zero
	if boundary.Nanosecond() != 0 {
		t.Errorf("expected nanoseconds=0, got %d", boundary.Nanosecond())
	}

	// Should not be in the future (in any timezone)
	if boundary.After(time.Now().Add(24 * time.Hour)) {
		t.Error("boundary should not be more than 24 hours in the future")
	}
}

func TestMakassarTimezone_Constant(t *testing.T) {
	// Verify the constant is what we expect
	if MakassarTimezone != "Asia/Makassar" {
		t.Errorf("expected 'Asia/Makassar', got '%s'", MakassarTimezone)
	}
}