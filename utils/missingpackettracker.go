package utils

import (
	"fmt"
	"sync"
)

// MissingPacketsTracker is used to track missing packets/flows
type MissingPacketsTracker struct {
	packetsCount            map[string]int64 // map[SOURCE_ADDR_DOMAIN_KEY]ACTUAL_FLOWS/PACKET_COUNT
	packetsLastMissingCount map[string]int64 // map[SOURCE_ADDR_DOMAIN_KEY]LAST_FLOWS/PACKET_MISSING_COUNT
	packetsCountMu          *sync.RWMutex

	maxNegativeSequenceDifference int
}

func NewMissingPacketsTracker(maxNegativeSequenceDifference int) *MissingPacketsTracker {
	return &MissingPacketsTracker{
		packetsCount:                  make(map[string]int64),
		packetsLastMissingCount:       make(map[string]int64),
		packetsCountMu:                &sync.RWMutex{},
		maxNegativeSequenceDifference: maxNegativeSequenceDifference,
	}
}

func (s *MissingPacketsTracker) countMissing(key string, seqnum uint32, packetOrFlowCount uint16) uint64 {
	s.packetsCountMu.Lock()
	defer s.packetsCountMu.Unlock()

	fmt.Printf("[countMissing] step1 seqnum=%d, packetsCount=%d, last=%d\n", seqnum, s.packetsCount[key], s.packetsLastMissingCount[key])
	if _, ok := s.packetsCount[key]; !ok {
		s.packetsCount[key] = int64(seqnum)
	} else {
		s.packetsCount[key] += int64(packetOrFlowCount)
	}
	totalMissing := int64(seqnum) - s.packetsCount[key]
	fmt.Printf("[countMissing] step2 seqnum=%d, packetsCount=%d, last=%d, totalMissing=%d\n", seqnum, s.packetsCount[key], s.packetsLastMissingCount[key], totalMissing)

	// There is likely a sequence number reset when the number of missing packets is negative and very high.
	// In this case, we save the current sequence number of consider that there is no missing packets.
	if totalMissing <= -int64(s.maxNegativeSequenceDifference) {
		s.packetsCount[key] = int64(seqnum)
		s.packetsLastMissingCount[key] = 0
		totalMissing = 0
	}
	newMissing := totalMissing - s.packetsLastMissingCount[key]
	fmt.Printf("[countMissing] step3 seqnum=%d, packetsCount=%d, last=%d, totalMissing=%d, newMissing=%d\n", seqnum, s.packetsCount[key], s.packetsLastMissingCount[key], totalMissing, newMissing)

	// ensure that we never return negative missing count
	if newMissing > 0 {
		s.packetsLastMissingCount[key] = totalMissing
		return uint64(newMissing)
	}
	return 0
}
