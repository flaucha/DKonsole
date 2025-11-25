package prometheus

// MetricDataPoint represents a single data point in a time series
type MetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// DeploymentMetricsResponse includes CPU and Memory metrics for a deployment
type DeploymentMetricsResponse struct {
	CPU    []MetricDataPoint `json:"cpu"`
	Memory []MetricDataPoint `json:"memory"`
}

// PodMetricsResponse includes all pod metrics
type PodMetricsResponse struct {
	CPU       []MetricDataPoint `json:"cpu"`
	Memory    []MetricDataPoint `json:"memory"`
	NetworkRx []MetricDataPoint `json:"networkRx"`
	NetworkTx []MetricDataPoint `json:"networkTx"`
	PVCUsage  []MetricDataPoint `json:"pvcUsage"`
}

// NodeMetric represents metrics for a single node
type NodeMetric struct {
	Name      string  `json:"name"`
	CPUUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memoryUsage"`
	DiskUsage float64 `json:"diskUsage"`
	NetworkRx float64 `json:"networkRx"`
	NetworkTx float64 `json:"networkTx"`
	Status    string  `json:"status"`
}

// ClusterStats represents aggregated cluster statistics from Prometheus
type ClusterStats struct {
	TotalNodes     int     `json:"totalNodes"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	NetworkTraffic float64 `json:"networkTraffic"`
	CPUTrend       float64 `json:"cpuTrend"`
	MemoryTrend    float64 `json:"memoryTrend"`
}

// ClusterOverviewResponse includes cluster-wide metrics
type ClusterOverviewResponse struct {
	NodeMetrics  []NodeMetric  `json:"nodeMetrics"`
	ClusterStats *ClusterStats `json:"clusterStats"`
}

// StatusResponse represents the Prometheus service status
type StatusResponse struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}









