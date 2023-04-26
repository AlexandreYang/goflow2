package utils

import (
	"sync"
)

// MissingFlowsTracker is used to track missing packets/flows
type MissingFlowsTracker struct {
	// counter/lastSequences key is based on source addr and sourceId/obsDomain/engineType/engineId
	counters                      map[string]int64
	lastSequences                 map[string]int64
	mutex                         *sync.RWMutex
	maxNegativeSequenceDifference int
}

func NewMissingFlowsTracker(maxNegativeSequenceDifference int) *MissingFlowsTracker {
	return &MissingFlowsTracker{
		counters:                      make(map[string]int64),
		lastSequences:                 make(map[string]int64),
		mutex:                         &sync.RWMutex{},
		maxNegativeSequenceDifference: maxNegativeSequenceDifference,
	}
}

func (s *MissingFlowsTracker) countMissing(key string, seqnum uint32, flows uint16) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.counters[key]; !ok {
		s.counters[key] = int64(seqnum)
	} else {
		s.counters[key] += int64(flows)
	}

	var missingElements int64
	// We assume there is a sequence number reset when the current sequence number
	// minus prev sequence number is negative and high.
	// When this happens, we reset the counter to the current sequence number.
	sequenceDiff := int64(seqnum) - s.lastSequences[key]
	if sequenceDiff <= -int64(s.maxNegativeSequenceDifference) {
		s.counters[key] = int64(seqnum)
		missingElements = 0
	} else {
		missingElements = int64(seqnum) - s.counters[key]
	}
	s.lastSequences[key] = int64(seqnum)
	return missingElements
}
