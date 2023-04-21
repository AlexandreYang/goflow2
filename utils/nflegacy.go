package utils

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/netsampler/goflow2/decoders/netflowlegacy"
	"github.com/netsampler/goflow2/format"
	flowmessage "github.com/netsampler/goflow2/pb"
	"github.com/netsampler/goflow2/producer"
	"github.com/netsampler/goflow2/transport"
	"github.com/prometheus/client_golang/prometheus"
)

type StateNFLegacy struct {
	stopper

	Format    format.FormatInterface
	Transport transport.TransportInterface
	Logger    Logger

	// sequenceTracker is used to track missing packets
	// structure: map[PACKET_SOURCE_ADDR]LAST_SEQUENCE_NUMBER
	sequenceTracker     map[string]int64
	sequenceTrackerPrev map[string]int64
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

		seqnum := msgDecConv.FlowSequence
		flowCount := int64(msgDecConv.Count)

		fmt.Printf("[GOFLOW] 1Sequence Number: %s - %d\n", samplerAddress.String(), seqnum)
		sequenceTrackerKey := samplerAddress.String()

		// TODO: More granular lock location?
		s.sequenceTrackerLock.Lock()
		if _, ok := s.sequenceTracker[sequenceTrackerKey]; !ok {
			s.sequenceTracker[sequenceTrackerKey] = int64(seqnum) - int64(flowCount)
			s.sequenceTrackerPrev[sequenceTrackerKey] = int64(seqnum) - int64(flowCount)
		}
		if _, ok := s.sequenceTracker[sequenceTrackerKey]; ok {
			prevTracked := s.sequenceTracker[sequenceTrackerKey]
			numFlows := int64(flowCount)
			fmt.Printf("[GOFLOW] 2Sequence Number: %s - last=%d, flows=%d, seqnum=%d\n", samplerAddress.String(), prevTracked, numFlows, seqnum)
			fmt.Printf("[GOFLOW] 3trackedFlows=%d\n", prevTracked)
			fmt.Printf("[GOFLOW] 4numFlows=%d\n", numFlows)
			fmt.Printf("[GOFLOW] 5seqnum=%d\n", seqnum)

			s.sequenceTracker[sequenceTrackerKey] = prevTracked + numFlows
			missingFlows := int64(seqnum) - s.sequenceTracker[sequenceTrackerKey]
			if missingFlows <= 10000 {
				s.sequenceTracker[sequenceTrackerKey] = int64(seqnum)
			}
			if missingFlows != 0 {
				fmt.Printf("[GOFLOW] 6Sequence Number: %s - last=%d, flows=%d, seqnum=%d : missing flows=%d\n", samplerAddress.String(), prevTracked, numFlows, seqnum, missingFlows)
			}
			missingFlowPrev := int64(seqnum) - (s.sequenceTrackerPrev[sequenceTrackerKey] + flowCount)
			if missingFlowPrev != 0 {
				fmt.Printf("[GOFLOW] 7Sequence Number: %s - last=%d, flows=%d, seqnum=%d : missing flows prev=%d\n", samplerAddress.String(), prevTracked, numFlows, seqnum, missingFlowPrev)
			}
			s.sequenceTrackerPrev[sequenceTrackerKey] = int64(seqnum)

			NetFlowSequenceDelta.With(
				prometheus.Labels{
					"router":  key,
					"version": "5",
				}).
				Set(float64(missingFlows))
		}
		s.sequenceTrackerLock.Unlock()

		// TODO: detect sequence number gap and send as prometheus metric
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

func (s *StateNFLegacy) initConfig() {
	s.sequenceTracker = make(map[string]int64)
	s.sequenceTrackerPrev = make(map[string]int64)
	s.sequenceTrackerLock = &sync.RWMutex{}
}

func (s *StateNFLegacy) FlowRoutine(workers int, addr string, port int, reuseport bool) error {
	if err := s.start(); err != nil {
		return err
	}
	s.initConfig()
	return UDPStoppableRoutine(s.stopCh, "NetFlowV5", s.DecodeFlow, workers, addr, port, reuseport, s.Logger)
}
