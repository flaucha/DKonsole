package models

// PodMetric representa métricas de CPU y memoria de un Pod
type PodMetric struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// PrometheusQueryResult representa el resultado de una consulta a Prometheus
type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// MetricDataPoint representa un punto de datos de métrica con timestamp y valor
type MetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// DeploymentMetricsResponse contiene métricas de CPU y memoria de un Deployment
type DeploymentMetricsResponse struct {
	CPU    []MetricDataPoint `json:"cpu"`
	Memory []MetricDataPoint `json:"memory"`
}

// PodMetricsResponse incluye todas las métricas de un Pod
type PodMetricsResponse struct {
	CPU       []MetricDataPoint `json:"cpu"`
	Memory    []MetricDataPoint `json:"memory"`
	NetworkRx []MetricDataPoint `json:"networkRx"`
	NetworkTx []MetricDataPoint `json:"networkTx"`
	PVCUsage  []MetricDataPoint `json:"pvcUsage"`
}

// ClusterOverviewResponse incluye métricas a nivel de cluster
type ClusterOverviewResponse struct {
	NodeMetrics  []NodeMetric            `json:"nodeMetrics"`
	ClusterStats *PrometheusClusterStats `json:"clusterStats"`
}

// NodeMetric representa métricas para un nodo individual
type NodeMetric struct {
	Name      string  `json:"name"`
	Role      string  `json:"role"` // "worker" or "control-plane"
	CPUUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memoryUsage"`
	DiskUsage float64 `json:"diskUsage"`
	NetworkRx float64 `json:"networkRx"`
	NetworkTx float64 `json:"networkTx"`
	Status    string  `json:"status"`
}

// PrometheusClusterStats representa estadísticas agregadas del cluster desde Prometheus
type PrometheusClusterStats struct {
	TotalNodes        int     `json:"totalNodes"`
	ControlPlaneNodes int     `json:"controlPlaneNodes"`
	AvgCPUUsage       float64 `json:"avgCpuUsage"`
	AvgMemoryUsage    float64 `json:"avgMemoryUsage"`
	NetworkTraffic    float64 `json:"networkTraffic"`
	CPUTrend          float64 `json:"cpuTrend"`
	MemoryTrend       float64 `json:"memoryTrend"`
}

// StatusResponse representa el estado del servicio Prometheus
type StatusResponse struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}
