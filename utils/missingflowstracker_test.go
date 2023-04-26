package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissingFlowsTracker_countMissingFlows(t *testing.T) {
	tests := []struct {
		name                       string
		sequenceTrackerKey         string
		seqnum                     uint32
		flowCount                  uint16
		savedCounters              map[string]int64
		savedLastSequences         map[string]int64
		expectedMissingFlows       int64
		expectedSavedCounters      map[string]int64
		expectedSavedLastSequences map[string]int64
	}{
		{
			name:                 "no saved seq tracker yet",
			savedCounters:        map[string]int64{},
			savedLastSequences:   map[string]int64{},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               100,
			flowCount:            100,
			expectedMissingFlows: 0,
			expectedSavedCounters: map[string]int64{
				"127.0.01": 100,
			},
			expectedSavedLastSequences: map[string]int64{
				"127.0.01": 100,
			},
		},
		{
			name: "no missing flows",
			savedCounters: map[string]int64{
				"127.0.01": 100,
			},
			savedLastSequences: map[string]int64{
				"127.0.01": 100,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               200,
			flowCount:            100,
			expectedMissingFlows: 0,
			expectedSavedCounters: map[string]int64{
				"127.0.01": 200,
			},
			expectedSavedLastSequences: map[string]int64{
				"127.0.01": 200,
			},
		},
		{
			name: "have missing flows",
			savedCounters: map[string]int64{
				"127.0.01": 100,
			},
			savedLastSequences: map[string]int64{
				"127.0.01": 100,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               200,
			flowCount:            30,
			expectedMissingFlows: 70,
			expectedSavedCounters: map[string]int64{
				"127.0.01": 130,
			},
			expectedSavedLastSequences: map[string]int64{
				"127.0.01": 200,
			},
		},
		{
			name: "negative saved sequence tracker",
			// reported missing flows count can be temporarily negative when udp packet arrive unordered,
			// when slightly lower sequence number arrives after higher sequence number.
			savedCounters: map[string]int64{
				"127.0.01": 1000,
			},
			savedLastSequences: map[string]int64{
				"127.0.01": 1000,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               950,
			flowCount:            10,
			expectedMissingFlows: -60,
			expectedSavedCounters: map[string]int64{
				"127.0.01": 1010,
			},
			expectedSavedLastSequences: map[string]int64{
				"127.0.01": 950,
			},
		},
		{
			name: "sequence number reset",
			savedCounters: map[string]int64{
				"127.0.01": 9000,
			},
			savedLastSequences: map[string]int64{
				"127.0.01": 9000,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               2000,
			flowCount:            100,
			expectedMissingFlows: 0,
			expectedSavedCounters: map[string]int64{
				"127.0.01": 2000,
			},
			expectedSavedLastSequences: map[string]int64{
				"127.0.01": 2000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMissingFlowsTracker(1000)
			s.counters = tt.savedCounters
			s.lastSequences = tt.savedLastSequences
			assert.Equal(t, tt.expectedMissingFlows, s.countMissing(tt.sequenceTrackerKey, tt.seqnum, tt.flowCount))
			assert.Equal(t, tt.expectedSavedCounters, s.counters)
			assert.Equal(t, tt.expectedSavedLastSequences, s.lastSequences)
		})
	}
}
