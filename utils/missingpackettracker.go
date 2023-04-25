package utils

import (
	"fmt"
	"sync"
)

// MissingPacketsTracker is used to track missing packets/flows
type MissingPacketsTracker struct {
	packetsCount   map[string]int64 // map[SOURCE_ADDR_DOMAIN_KEY]ACTUAL_FLOWS/PACKET_COUNT
	packetsCountMu *sync.RWMutex

	maxNegativeSequenceDifference int
}

func NewMissingPacketsTracker(maxNegativeSequenceDifference int) *MissingPacketsTracker {
	return &MissingPacketsTracker{
		packetsCount:                  make(map[string]int64),
		packetsCountMu:                &sync.RWMutex{},
		maxNegativeSequenceDifference: maxNegativeSequenceDifference,
	}
}

func (s *MissingPacketsTracker) countMissing(key string, seqnum uint32, packetOrFlowCount uint16) int64 {
	s.packetsCountMu.Lock()
	defer s.packetsCountMu.Unlock()

	fmt.Printf("[countMissing] step1 seqnum=%d, packetsCount=%d", seqnum, s.packetsCount[key])
	if _, ok := s.packetsCount[key]; !ok {
		s.packetsCount[key] = int64(seqnum)
	} else {
		s.packetsCount[key] += int64(packetOrFlowCount)
	}
	missingElements := int64(seqnum) - s.packetsCount[key]
	fmt.Printf("[countMissing] step2 seqnum=%d, packetsCount=%d, missingElements=%d", seqnum, s.packetsCount[key], missingElements)

	// There is likely a sequence number reset when the number of missing packets is negative and very high.
	// In this case, we save the current sequence number of consider that there is no missing packets.
	if missingElements <= -int64(s.maxNegativeSequenceDifference) {
		s.packetsCount[key] = int64(seqnum)
		missingElements = 0
	}
	return missingElements
}
