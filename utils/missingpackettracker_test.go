package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissingFlowsTracker_countMissingFlows(t *testing.T) {
	tests := []struct {
		name                                 string
		sequenceTrackerKey                   string
		seqnum                               uint32
		flowCount                            uint16
		savedPacketsCount                    map[string]int64
		savedPacketsLastMissingCount         map[string]int64
		expectedMissingFlows                 uint64
		expectedSavedPacketsCount            map[string]int64
		expectedSavedPacketsLastMissingCount map[string]int64
	}{
		{
			name:                         "no saved seq tracker yet",
			savedPacketsCount:            map[string]int64{},
			savedPacketsLastMissingCount: map[string]int64{},
			sequenceTrackerKey:           "127.0.01",
			seqnum:                       100,
			flowCount:                    100,
			expectedMissingFlows:         0,
			expectedSavedPacketsCount: map[string]int64{
				"127.0.01": 100,
			},
			expectedSavedPacketsLastMissingCount: map[string]int64{},
		},
		{
			name: "no missing flows",
			savedPacketsCount: map[string]int64{
				"127.0.01": 100,
			},
			savedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               200,
			flowCount:            100,
			expectedMissingFlows: 0,
			expectedSavedPacketsCount: map[string]int64{
				"127.0.01": 200,
			},
			expectedSavedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
		},
		{
			name: "have missing flows",
			savedPacketsCount: map[string]int64{
				"127.0.01": 100,
			},
			savedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               200,
			flowCount:            30,
			expectedMissingFlows: 70,
			expectedSavedPacketsCount: map[string]int64{
				"127.0.01": 130,
			},
			expectedSavedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 70,
			},
		},
		{
			name: "negative saved sequence tracker",
			// reported missing flows count can be temporarily negative when udp packet arrive unordered,
			// when slightly lower sequence number arrives after higher sequence number.
			savedPacketsCount: map[string]int64{
				"127.0.01": 1000,
			},
			savedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               950,
			flowCount:            10,
			expectedMissingFlows: 0,
			expectedSavedPacketsCount: map[string]int64{
				"127.0.01": 1010,
			},
			expectedSavedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
		},
		{
			name: "sequence number reset",
			savedPacketsCount: map[string]int64{
				"127.0.01": 9000,
			},
			savedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 100,
			},
			sequenceTrackerKey:   "127.0.01",
			seqnum:               2000,
			flowCount:            100,
			expectedMissingFlows: 0,
			expectedSavedPacketsCount: map[string]int64{
				"127.0.01": 2000,
			},
			expectedSavedPacketsLastMissingCount: map[string]int64{
				"127.0.01": 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMissingPacketsTracker(1000)
			s.packetsCount = tt.savedPacketsCount
			s.packetsLastMissingCount = tt.savedPacketsLastMissingCount
			assert.Equal(t, tt.expectedMissingFlows, s.countMissing(tt.sequenceTrackerKey, tt.seqnum, tt.flowCount))
			assert.Equal(t, tt.expectedSavedPacketsCount, s.packetsCount)
			assert.Equal(t, tt.expectedSavedPacketsLastMissingCount, s.packetsLastMissingCount)
		})
	}
}
