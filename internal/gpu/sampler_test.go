package gpu

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const bootId = "boot-123"
const tsString = "2026-08-06 21:00:00"

// Testing gap: unable to test the EBUSY sleeping-card skip, but live verification has been proven that the skip works
func TestSample(t *testing.T) {
	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	s, err := NewSampler("testdata/drm", bootId, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error creating sampler: %v", err)
	}

	wantCardDirs := []string{"testdata/drm/card1", "testdata/drm/card2"}

	gotCardDirs, err := s.detectCards()
	if err != nil {
		t.Fatalf("unexpected error detecting cards: %v", err)
	}

	if match := reflect.DeepEqual(gotCardDirs, wantCardDirs); !match {
		t.Errorf("expected match. wanted '%+v', but got '%+v'", wantCardDirs, gotCardDirs)
	}

	gotCard1Sample, err := s.sampleCard(gotCardDirs[0], ts)
	if err != nil {
		t.Fatalf("unexpected error sampling card1: %v", err)
	}

	wantCard1Sample := Entry{
		BootId:       bootId,
		Ts:           ts,
		Card:         "card1",
		BusyPercent:  int64(42),
		VramUsed:     int64(1073741824),
		VramTotal:    int64(8589934592),
		GttUsed:      int64(268425456),
		TempEdge:     new(int64(45)),
		TempJunction: new(int64(52)),
		TempMem:      new(int64(60)),
		PowerWatts:   new(int64(87)),
		FanRpm:       new(int64(1450)),
		SclkMhz:      new(int64(2100)),
		MclkMhz:      new(int64(1750)),
	}

	if diff := cmp.Diff(wantCard1Sample, gotCard1Sample); diff != "" {
		t.Errorf("Parse mismatch card1: %v", diff)
	}

	gotCard2Sample, err := s.sampleCard(gotCardDirs[1], ts)
	if err != nil {
		t.Fatalf("unexpected error sampling card2: %v", err)
	}

	wantCard2Sample := Entry{
		BootId:      bootId,
		Ts:          ts,
		Card:        "card2",
		BusyPercent: int64(7),
		VramUsed:    int64(536870912),
		VramTotal:   int64(4294967296),
		GttUsed:     int64(134217728),
		TempEdge:    new(int64(55)),
		PowerWatts:  new(int64(15)),
		SclkMhz:     new(int64(8000)),
	}

	if diff := cmp.Diff(wantCard2Sample, gotCard2Sample); diff != "" {
		t.Errorf("Parse mismatch card2: %v", diff)
	}
}

func TestNoCards(t *testing.T) {
	s, err := NewSampler(t.TempDir(), bootId, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error creating sampler: %v", err)
	}

	if _, err := s.detectCards(); err == nil {
		t.Errorf("expected an error creating sampler, but got none")
	}
}

func TestFullSample(t *testing.T) {
	s, err := NewSampler("testdata/drm", bootId, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error creating sampler: %v", err)
	}

	gotCardDirs, err := s.detectCards()
	if err != nil {
		t.Fatalf("unexpected error fetching card directories: %v", err)
	}

	events, err := s.Sample(context.Background())
	if err != nil {
		t.Fatalf("unexpected error sampling: %v", err)
	}

	countCards := make(map[string]int)
	for _, cardDir := range gotCardDirs {
		countCards[cardDir]++
	}

	for _, want := range []string{"testdata/drm/card1", "testdata/drm/card2"} {
		if countCards[want] != 1 {
			t.Errorf("%s has %d entries, want 1", want, countCards[want])
		}
	}

	if len(events) != 2 {
		t.Errorf("expected exactly 2 events, but got %d", len(events))
	}

}
