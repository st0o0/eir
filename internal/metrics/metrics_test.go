package metrics

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordEvent(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := New(reg)

	m.RecordEvent("die")
	m.RecordEvent("die")
	m.RecordEvent("start")

	expected := `
# HELP eir_events_received_total Total number of Docker events received.
# TYPE eir_events_received_total counter
eir_events_received_total{action="die"} 2
eir_events_received_total{action="start"} 1
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "eir_events_received_total"); err != nil {
		t.Error(err)
	}
}

func TestRecordHeal_Success(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := New(reg)

	m.RecordHeal("bifrost", nil, 3*time.Second)

	expected := `
# HELP eir_heals_total Total number of healing operations.
# TYPE eir_heals_total counter
eir_heals_total{master="bifrost",status="success"} 1
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "eir_heals_total"); err != nil {
		t.Error(err)
	}
}

func TestRecordHeal_Failure(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := New(reg)

	m.RecordHeal("bifrost", errors.New("timeout"), 5*time.Second)

	expected := `
# HELP eir_heals_total Total number of healing operations.
# TYPE eir_heals_total counter
eir_heals_total{master="bifrost",status="fail"} 1
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "eir_heals_total"); err != nil {
		t.Error(err)
	}
}
