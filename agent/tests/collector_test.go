package tests

import (
	"context"
	"testing"
	"time"

	"github.com/havilz/caelus-cloud/agent/internal/collector"
)

func TestCollector_CollectSuccess(t *testing.T) {
	c := collector.NewCollector()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("expected no error during collection, got: %v", err)
	}

	if metrics == nil {
		t.Fatal("expected non-nil metrics result")
	}

	if metrics.CPUCores <= 0 {
		t.Errorf("expected CPUCores > 0, got %d", metrics.CPUCores)
	}

	if metrics.MemoryTotalMB <= 0 {
		t.Errorf("expected MemoryTotalMB > 0, got %d", metrics.MemoryTotalMB)
	}

	if metrics.DiskTotalGB <= 0 {
		t.Errorf("expected DiskTotalGB > 0, got %f", metrics.DiskTotalGB)
	}

	if metrics.OS == "" {
		t.Errorf("expected non-empty OS name")
	}

	if metrics.Platform == "" {
		t.Errorf("expected non-empty Platform string")
	}
}

func TestCollector_ConsecutiveDeltaCalculation(t *testing.T) {
	c := collector.NewCollector()
	ctx := context.Background()

	m1, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("first collection failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	m2, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("second collection failed: %v", err)
	}

	if m1 == nil || m2 == nil {
		t.Fatal("expected valid metrics on both calls")
	}

	if m2.CPUUsagePct < 0.0 || m2.CPUUsagePct > 100.0 {
		t.Errorf("CPU usage percentage out of bounds: %f", m2.CPUUsagePct)
	}

	if m2.MemoryUsagePct < 0.0 || m2.MemoryUsagePct > 100.0 {
		t.Errorf("Memory usage percentage out of bounds: %f", m2.MemoryUsagePct)
	}
}
