package utils

import (
	"bytes"
	"sync"
	"time"

	"github.com/netsampler/goflow2/decoders/netflowlegacy"
	"github.com/netsampler/goflow2/format"
	flowmessage "github.com/netsampler/goflow2/pb"
	"github.com/netsampler/goflow2/producer"
	"github.com/netsampler/goflow2/transport"
	"github.com/prometheus/client_golang/prometheus"
)

var MaxNegativeSequenceDifference = 1000

type StateNFLegacy struct {
	stopper

	Format    format.FormatInterface
	Transport transport.TransportInterface
	Logger    Logger

	// savedSeqTracker is used to track missing packets
	// structure: map[PACKET_SOURCE_ADDR]LAST_SEQUENCE_NUMBER
	savedSeqTracker     map[string]int64
	sequenceTrackerLock *sync.RWMutex
}

func (s *StateNFLegacy) DecodeFlow(msg interface{}) error {
	pkt := msg.(BaseMessage)
	buf := bytes.NewBuffer(pkt.Payload)
	key := pkt.Src.String()
	samplerAddress := pkt.Src
	if samplerAddress.To4() != nil {
		samplerAddress = samplerAddress.To4()
	}

	ts := uint64(time.Now().UTC().Unix())
	if pkt.SetTime {
		ts = uint64(pkt.RecvTime.UTC().Unix())
	}

	timeTrackStart := time.Now()
	msgDec, err := netflowlegacy.DecodeMessage(buf)

	if err != nil {
		switch err.(type) {
		case *netflowlegacy.ErrorVersion:
			NetFlowErrors.With(
				prometheus.Labels{
					"router": key,
					"error":  "error_version",
				}).
				Inc()
		}
		return err
	}

	switch msgDecConv := msgDec.(type) {
	case netflowlegacy.PacketNetFlowV5:
		NetFlowStats.With(
			prometheus.Labels{
				"router":  key,
				"version": "5",
			}).
			Inc()
		NetFlowSetStatsSum.With(
			prometheus.Labels{
				"router":  key,
				"version": "5",
				"type":    "DataFlowSet",
			}).
			Add(float64(msgDecConv.Count))

		missingFlows := s.countMissingFlows(samplerAddress.String(), msgDecConv.FlowSequence, msgDecConv.Count)

		NetFlowSequenceDelta.With(
			prometheus.Labels{
				"router":  key,
				"version": "5",
			}).
			Set(float64(missingFlows))
	}

	var flowMessageSet []*flowmessage.FlowMessage
	flowMessageSet, err = producer.ProcessMessageNetFlowLegacy(msgDec)

	timeTrackStop := time.Now()
	DecoderTime.With(
		prometheus.Labels{
			"name": "NetFlowV5",
		}).
		Observe(float64((timeTrackStop.Sub(timeTrackStart)).Nanoseconds()) / 1000)

	for _, fmsg := range flowMessageSet {
		fmsg.TimeReceived = ts
		fmsg.SamplerAddress = samplerAddress

		if s.Format != nil {
			key, data, err := s.Format.Format(fmsg)

			if err != nil && s.Logger != nil {
				s.Logger.Error(err)
			}
			if err == nil && s.Transport != nil {
				s.Transport.Send(key, data)
			}
		}
	}

	return nil
}

func (s *StateNFLegacy) countMissingFlows(sequenceTrackerKey string, seqnum uint32, flowCount uint16) int64 {
	s.sequenceTrackerLock.Lock()
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

	s.sequenceTrackerLock.Unlock()
	return missingFlows
}

func (s *StateNFLegacy) initConfig() {
	s.savedSeqTracker = make(map[string]int64)
	s.sequenceTrackerLock = &sync.RWMutex{}
}

func (s *StateNFLegacy) FlowRoutine(workers int, addr string, port int, reuseport bool) error {
	if err := s.start(); err != nil {
		return err
	}
	s.initConfig()
	return UDPStoppableRoutine(s.stopCh, "NetFlowV5", s.DecodeFlow, workers, addr, port, reuseport, s.Logger)
}
