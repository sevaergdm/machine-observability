package gpu

import (
	"context"
	"fmt"
	"log/slog"
	"machine-observability/internal/collector"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Sampler struct {
	BootId   string
	Root     string
	Logger   *slog.Logger
	CardDirs []string
}

func NewSampler(root, bootId string, logger *slog.Logger) (*Sampler, error) {
	s := &Sampler{BootId: bootId, Root: root, Logger: logger}
	cards, err := s.detectCards()
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("no amdgpu cards under %s", root)
	}
	s.CardDirs = cards
	return s, nil
}

func (s *Sampler) Sample(ctx context.Context) ([]collector.Event, error) {
	var events []collector.Event
	for _, cardDir := range s.CardDirs {
		e, err := s.sampleCards(cardDir, time.Now().UTC())
		if err != nil {
			s.Logger.Error("unable to sample card", "error", err, "card", cardDir)
			continue
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Sampler) sampleCards(cardDir string, ts time.Time) (Entry, error) {
	cardName := strings.TrimPrefix(cardDir, s.Root+"/")
	dev := filepath.Join(cardDir, "device")

	busy, err := readInt(filepath.Join(dev, "gpu_busy_percent"))
	if err != nil {
		return Entry{}, err
	}

	vramUsed, err := readInt(filepath.Join(dev, "mem_info_vram_used"))
	if err != nil {
		return Entry{}, err
	}

	vramTotal, err := readInt(filepath.Join(dev, "mem_info_vram_total"))
	if err != nil {
		return Entry{}, err
	}

	gttUsed, err := readInt(filepath.Join(dev, "mem_info_gtt_used"))
	if err != nil {
		return Entry{}, err
	}

	hw := findHwmonDir(dev)

	return Entry{
		BootId:       s.BootId,
		Ts:           ts,
		Card:         cardName,
		BusyPercent:  busy,
		VramUsed:     vramUsed,
		VramTotal:    vramTotal,
		GttUsed:      gttUsed,
		TempEdge:     readOptInt(filepath.Join(hw, "temp1_input"), 1000),
		TempJunction: readOptInt(filepath.Join(hw, "temp2_input"), 1000),
		TempMem:      readOptInt(filepath.Join(hw, "temp3_input"), 1000),
		PowerWatts:   readOptInt(filepath.Join(hw, "power1_average"), 1_000_000),
		FanRpm:       readOptInt(filepath.Join(hw, "fan1_input"), 1),
		SclkMhz:      readOptInt(filepath.Join(hw, "freq1_input"), 1_000_000),
		MclkMhz:      readOptInt(filepath.Join(hw, "freq2_input"), 1_000_000),
	}, nil
}

func (s *Sampler) detectCards() ([]string, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", s.Root, err)
	}

	var cards []string
	for _, entry := range entries {
		rest, ok := strings.CutPrefix(entry.Name(), "card")
		if !ok || rest == "" {
			continue
		}
		if _, err := strconv.Atoi(rest); err != nil {
			continue
		}

		if _, err := os.Stat(filepath.Join(s.Root, entry.Name(), "device", "gpu_busy_percent")); err != nil {
			continue
		}
		cards = append(cards, filepath.Join(s.Root, entry.Name()))
	}
	return cards, nil
}

func findHwmonDir(path string) string {
	matches, err := filepath.Glob(filepath.Join(path, "hwmon", "hwmon*"))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func readInt(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}

	return v, nil
}

func readOptInt(path string, div int64) *int64 {
	v, err := readInt(path)
	if err != nil {
		return nil
	}
	v /= div
	return &v
}
