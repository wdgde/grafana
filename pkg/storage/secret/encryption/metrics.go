package encryption

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "grafana_secrets_manager"
	subsystem = "storage"
)

// DataKeyMetrics is a struct that contains all the metrics for all operations of encryption storage.
type DataKeyMetrics struct {
	CreateDataKeyDuration     prometheus.Histogram
	GetDataKeyDuration        prometheus.Histogram
	GetCurrentDataKeyDuration prometheus.Histogram
	ListDataKeysDuration      prometheus.Histogram
	DisableDataKeysDuration   prometheus.Histogram
	DeleteDataKeyDuration     prometheus.Histogram
}

func newDataKeyMetrics() *DataKeyMetrics {
	return &DataKeyMetrics{
		CreateDataKeyDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_create_duration_seconds",
			Help:      "Duration of create data key operations",
			Buckets:   prometheus.DefBuckets,
		}),
		GetDataKeyDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_get_duration_seconds",
			Help:      "Duration of get data key operations",
			Buckets:   prometheus.DefBuckets,
		}),
		GetCurrentDataKeyDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_get_current_duration_seconds",
			Help:      "Duration of get current data key operations",
			Buckets:   prometheus.DefBuckets,
		}),
		ListDataKeysDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_keys_list_duration_seconds",
			Help:      "Duration of list data keys operations",
			Buckets:   prometheus.DefBuckets,
		}),
		DisableDataKeysDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_keys_disable_duration_seconds",
			Help:      "Duration of disable data keys operations",
			Buckets:   prometheus.DefBuckets,
		}),
		DeleteDataKeyDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_delete_duration_seconds",
			Help:      "Duration of delete data key operations",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}

// NewDataKeyMetrics returns an instance of the DataKeyMetrics
// struct containing registered metrics if [reg] is not nil.
func NewDataKeyMetrics(reg prometheus.Registerer) *DataKeyMetrics {
	m := newDataKeyMetrics()

	if reg != nil {
		reg.MustRegister(
			m.CreateDataKeyDuration,
			m.GetDataKeyDuration,
			m.GetCurrentDataKeyDuration,
			m.ListDataKeysDuration,
			m.DisableDataKeysDuration,
			m.DeleteDataKeyDuration,
		)
	}

	return m
}

type GlobalDataKeyMetrics struct {
	DisableAllDataKeysDuration prometheus.Histogram
}

func newGlobalDataKeyMetrics() *GlobalDataKeyMetrics {
	return &GlobalDataKeyMetrics{

		DisableAllDataKeysDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disable_all_data_keys_duration_seconds",
			Help:      "Duration of disable all data keys operations",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}

// NewGlobalDataKeyMetrics returns an instance of the GlobalDataKeyMetrics
// struct containing registered metrics if [reg] is not nil.
func NewGlobalDataKeyMetrics(reg prometheus.Registerer) *GlobalDataKeyMetrics {
	m := newGlobalDataKeyMetrics()

	if reg != nil {
		reg.MustRegister(
			m.DisableAllDataKeysDuration,
		)
	}

	return m
}
