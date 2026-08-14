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
	BootId string
	Root   string
	Logger *slog.Logger
}

func (s *Sampler) Name() string { return "gpu" }

func NewSampler(root, bootId string, logger *slog.Logger) (*Sampler, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &Sampler{BootId: bootId, Root: root, Logger: logger}, nil
}

func (s *Sampler) Sample(ctx context.Context) ([]collector.Event, error) {
	var events []collector.Event
	ts := time.Now().UTC()

	cardDirs, err := s.detectCards()
	if err != nil {
		s.Logger.Error("unable to detect cards", "error", err)
		return nil, err
	}

	if len(cardDirs) == 0 {
		s.Logger.Error("no amdgpu cards found", "root", s.Root)
		return nil, err
	}

	for _, cardDir := range cardDirs {
		e, err := s.sampleCard(cardDir, ts)
		if err != nil {
			s.Logger.Debug("unable to sample card", "error", err, "card", cardDir)
			continue
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Sampler) sampleCard(cardDir string, ts time.Time) (Entry, error) {
	cardName := filepath.Base(cardDir)
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

	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards amdgpu cards detected")
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
