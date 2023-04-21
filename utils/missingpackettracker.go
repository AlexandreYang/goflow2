package utils

import (
	"sync"
)

// MissingPacketsTracker is used to track missing packets
type MissingPacketsTracker struct {
	packetsCount   map[string]int64 // map[SOURCE_ADDR_DOMAIN_KEY]ACTUAL_FLOWS_COUNT
	packetsCountMu *sync.RWMutex

	maxNegativeSequenceDifference int
}

func NewMissingFlowsTracker(maxNegativeSequenceDifference int) *MissingPacketsTracker {
	return &MissingPacketsTracker{
		packetsCount:                  make(map[string]int64),
		packetsCountMu:                &sync.RWMutex{},
		maxNegativeSequenceDifference: maxNegativeSequenceDifference,
	}
}

func (s *MissingPacketsTracker) countMissingPackets(sequenceTrackerKey string, seqnum uint32, packetCount uint16) int64 {
	s.packetsCountMu.Lock()
	defer s.packetsCountMu.Unlock()

	if _, ok := s.packetsCount[sequenceTrackerKey]; !ok {
		s.packetsCount[sequenceTrackerKey] = int64(seqnum)
	} else {
		s.packetsCount[sequenceTrackerKey] += int64(packetCount)
	}
	missingFlows := int64(seqnum) - s.packetsCount[sequenceTrackerKey]

	// There is likely a sequence number reset when the number of missing packets is negative and very high.
	// In this case, we save the current sequence number of consider that there is no missing packets.
	if missingFlows <= -int64(MaxNegativeFlowsSequenceDifference) {
		s.packetsCount[sequenceTrackerKey] = int64(seqnum)
		missingFlows = 0
	}
	return missingFlows
}
