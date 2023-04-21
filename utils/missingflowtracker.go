package utils

import (
	"sync"
)

type MissingFlowsTracker struct {
	// savedSeqTracker is used to track missing packets
	// structure: map[PACKET_SOURCE_ADDR]ACTUAL_FLOWS_COUNT
	savedSeqTracker     map[string]int64
	sequenceTrackerLock *sync.RWMutex
}

func NewMissingFlowsTracker() *MissingFlowsTracker {
	return &MissingFlowsTracker{
		savedSeqTracker:     make(map[string]int64),
		sequenceTrackerLock: &sync.RWMutex{},
	}
}

func (s *MissingFlowsTracker) countMissingFlows(sequenceTrackerKey string, seqnum uint32, flowCount uint16) int64 {
	s.sequenceTrackerLock.Lock()
	defer s.sequenceTrackerLock.Unlock()

	if _, ok := s.savedSeqTracker[sequenceTrackerKey]; !ok {
		s.savedSeqTracker[sequenceTrackerKey] = int64(seqnum)
	} else {
		s.savedSeqTracker[sequenceTrackerKey] += int64(flowCount)
	}
	missingFlows := int64(seqnum) - s.savedSeqTracker[sequenceTrackerKey]

	// There is likely a sequence number reset when the number of missing flows is negative and very high.
	// In this case, we save the current sequence number of consider that there is no missing flows.
	if missingFlows <= -int64(MaxNegativeSequenceDifference) {
		s.savedSeqTracker[sequenceTrackerKey] = int64(seqnum)
		missingFlows = 0
	}
	return missingFlows
}
