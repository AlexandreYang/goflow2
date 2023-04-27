package utils

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	MetricTrafficBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_traffic_bytes",
			Help: "Bytes received by the application.",
		},
		[]string{"remote_ip", "local_ip", "local_port", "type"},
	)
	MetricTrafficPackets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_traffic_packets",
			Help: "Packets received by the application.",
		},
		[]string{"remote_ip", "local_ip", "local_port", "type"},
	)
	MetricPacketSizeSum = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "flow_traffic_summary_size_bytes",
			Help:       "Summary of packet size.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"remote_ip", "local_ip", "local_port", "type"},
	)
	DecoderStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_decoder_count",
			Help: "Decoder processed count.",
		},
		[]string{"worker", "name"},
	)
	DecoderErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_decoder_error_count",
			Help: "Decoder processed error count.",
		},
		[]string{"worker", "name"},
	)
	DecoderTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "flow_summary_decoding_time_us",
			Help:       "Decoding time summary.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"name"},
	)
	DecoderProcessTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "flow_summary_processing_time_us",
			Help:       "Processing time summary.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"name"},
	)
	NetFlowStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_count",
			Help: "NetFlows processed.",
		},
		[]string{"router", "version"},
	)
	NetFlowErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_errors_count",
			Help: "NetFlows processed errors.",
		},
		[]string{"router", "error"},
	)
	NetFlowSetRecordsStatsSum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_flowset_records_sum",
			Help: "NetFlows FlowSets sum of records.",
		},
		[]string{"router", "version", "type"}, // data-template, data, opts...
	)
	NetFlowSetStatsSum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_flowset_sum",
			Help: "NetFlows FlowSets sum.",
		},
		[]string{"router", "version", "type"}, // data-template, data, opts...
	)
	NetFlowTimeStatsSum = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "flow_process_nf_delay_summary_seconds",
			Help:       "NetFlows time difference between time of flow and processing.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"router", "version"},
	)
	NetFlowTemplatesStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_templates_count",
			Help: "NetFlows Template count.",
		},
		[]string{"router", "version", "obs_domain_id", "template_id", "type"}, // options/template
	)
	NetFlowFlowsMissing = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_nf_flows_missing",
			Help: "NetFlows missing flows (mostly monotonic, but not always when packets arrive in unordered sequence).",
		},
		[]string{"router", "version", "engine_id", "engine_type"},
	)
	NetFlowFlowsSequence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_nf_flows_sequence",
			Help: "NetFlows last sequence number.",
		},
		[]string{"router", "version", "engine_id", "engine_type"},
	)
	NetFlowFlowsSequenceReset = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_flows_sequence_reset_count",
			Help: "NetFlows sequence reset count.",
		},
		[]string{"router", "version", "engine_id", "engine_type"},
	)
	NetFlowPacketsMissing = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_nf_packets_missing",
			Help: "NetFlows missing packets (mostly monotonic, but not always when packets in arrive unordered sequence).",
		},
		[]string{"router", "version", "obs_domain_id"},
	)
	NetFlowPacketsSequence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_nf_packets_sequence",
			Help: "NetFlows last sequence number.",
		},
		[]string{"router", "version", "obs_domain_id"},
	)
	NetFlowPacketsSequenceReset = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_nf_packets_sequence_reset_count",
			Help: "NetFlows sequence reset count.",
		},
		[]string{"router", "version", "obs_domain_id"},
	)
	SFlowStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_sf_count",
			Help: "sFlows processed.",
		},
		[]string{"router", "agent", "version"},
	)
	SFlowErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_sf_errors_count",
			Help: "sFlows processed errors.",
		},
		[]string{"router", "error"},
	)
	SFlowSampleStatsSum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_sf_samples_sum",
			Help: "SFlows samples sum.",
		},
		[]string{"router", "agent", "version", "type"}, // counter, flow, expanded...
	)
	SFlowSampleRecordsStatsSum = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_sf_samples_records_sum",
			Help: "SFlows samples sum of records.",
		},
		[]string{"router", "agent", "version", "type"}, // data-template, data, opts...
	)
	SFlowSamplesMissing = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_sf_samples_missing",
			Help: "SFlows missing flows (mostly monotonic, but not always when packets in arrive unordered sequence).",
		},
		[]string{"router", "version", "agent", "sub_agent_id"},
	)
	SFlowSamplesSequence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flow_process_sf_samples_sequence",
			Help: "SFlows last sequence number.",
		},
		[]string{"router", "version", "agent", "sub_agent_id"},
	)
	SFlowSamplesSequenceReset = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_process_sf_samples_sequence_reset_count",
			Help: "SFlows sequence reset count.",
		},
		[]string{"router", "version", "agent", "sub_agent_id"},
	)
)

func init() {
	prometheus.MustRegister(MetricTrafficBytes)
	prometheus.MustRegister(MetricTrafficPackets)
	prometheus.MustRegister(MetricPacketSizeSum)

	prometheus.MustRegister(DecoderStats)
	prometheus.MustRegister(DecoderErrors)
	prometheus.MustRegister(DecoderTime)
	prometheus.MustRegister(DecoderProcessTime)

	prometheus.MustRegister(NetFlowStats)
	prometheus.MustRegister(NetFlowErrors)
	prometheus.MustRegister(NetFlowSetRecordsStatsSum)
	prometheus.MustRegister(NetFlowSetStatsSum)
	prometheus.MustRegister(NetFlowTimeStatsSum)
	prometheus.MustRegister(NetFlowTemplatesStats)
	prometheus.MustRegister(NetFlowFlowsMissing)
	prometheus.MustRegister(NetFlowFlowsSequence)
	prometheus.MustRegister(NetFlowFlowsSequenceReset)
	prometheus.MustRegister(NetFlowPacketsMissing)
	prometheus.MustRegister(NetFlowPacketsSequence)
	prometheus.MustRegister(NetFlowPacketsSequenceReset)

	prometheus.MustRegister(SFlowStats)
	prometheus.MustRegister(SFlowErrors)
	prometheus.MustRegister(SFlowSampleStatsSum)
	prometheus.MustRegister(SFlowSampleRecordsStatsSum)
	prometheus.MustRegister(SFlowSamplesMissing)
	prometheus.MustRegister(SFlowSamplesSequence)
	prometheus.MustRegister(SFlowSamplesSequenceReset)
}

func DefaultAccountCallback(name string, id int, start, end time.Time) {
	DecoderProcessTime.With(
		prometheus.Labels{
			"name": name,
		}).
		Observe(float64((end.Sub(start)).Nanoseconds()) / 1000)
	DecoderStats.With(
		prometheus.Labels{
			"worker": strconv.Itoa(id),
			"name":   name,
		}).
		Inc()
}
