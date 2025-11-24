import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  fetchWorkloads,
  fetchPodLogs,
  fetchResource,
  fetchResources,
  fetchClusterOverview,
  fetchPrometheusStatus,
  fetchClusterMetrics,
  fetchNamespaces,
  fetchHelmReleases
} from '../k8sApi'

describe('k8sApi', () => {
  let mockFetcher

  beforeEach(() => {
    mockFetcher = vi.fn()
  })

  describe('fetchWorkloads', () => {
    it('should fetch workloads successfully', async () => {
      const mockData = [{ name: 'pod1', kind: 'Pod' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchWorkloads(mockFetcher, 'default', 'Pod')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/resources?kind=Pod&namespace=default')
    })

    it('should throw error when kind is missing', async () => {
      await expect(fetchWorkloads(mockFetcher, 'default', null)).rejects.toThrow('Resource kind is required')
      await expect(fetchWorkloads(mockFetcher, 'default', undefined)).rejects.toThrow('Resource kind is required')
    })

    it('should throw error when namespace is missing', async () => {
      await expect(fetchWorkloads(mockFetcher, null, 'Pod')).rejects.toThrow('Namespace is required')
      await expect(fetchWorkloads(mockFetcher, undefined, 'Pod')).rejects.toThrow('Namespace is required')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false,
        text: async () => 'Not Found'
      })

      await expect(fetchWorkloads(mockFetcher, 'default', 'Pod')).rejects.toThrow('Failed to fetch Pod: Not Found')
    })

    it('should use default fetch when fetcher is not provided', async () => {
      const originalFetch = global.fetch
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => []
      })

      await fetchWorkloads(null, 'default', 'Pod')
      expect(global.fetch).toHaveBeenCalled()

      global.fetch = originalFetch
    })
  })

  describe('fetchPodLogs', () => {
    it('should fetch pod logs successfully', async () => {
      const mockStream = new ReadableStream()
      mockFetcher.mockResolvedValue({
        ok: true,
        body: mockStream
      })

      const result = await fetchPodLogs(mockFetcher, 'default', 'pod1', 'container1')
      expect(result).toBe(mockStream)
      expect(mockFetcher).toHaveBeenCalledWith('/api/pods/logs?namespace=default&pod=pod1&container=container1')
    })

    it('should handle empty container name', async () => {
      const mockStream = new ReadableStream()
      mockFetcher.mockResolvedValue({
        ok: true,
        body: mockStream
      })

      await fetchPodLogs(mockFetcher, 'default', 'pod1', '')
      expect(mockFetcher).toHaveBeenCalledWith('/api/pods/logs?namespace=default&pod=pod1&container=')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchPodLogs(mockFetcher, 'default', 'pod1', 'container1')).rejects.toThrow('Failed to fetch logs')
    })
  })

  describe('fetchResource', () => {
    it('should fetch namespaced resource successfully', async () => {
      const mockData = { name: 'deployment1', kind: 'Deployment' }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchResource(mockFetcher, 'Deployment', 'deployment1', 'default')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/namespaces/default/Deployment/deployment1')
    })

    it('should fetch cluster-scoped resource successfully', async () => {
      const mockData = { name: 'node1', kind: 'Node' }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchResource(mockFetcher, 'Node', 'node1', null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/cluster/Node/node1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchResource(mockFetcher, 'Deployment', 'deployment1', 'default')).rejects.toThrow('Failed to fetch Deployment deployment1')
    })
  })

  describe('fetchResources', () => {
    it('should fetch resources with namespace', async () => {
      const mockData = [{ name: 'pod1' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchResources(mockFetcher, 'Pod', 'default', null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/resources?kind=Pod&namespace=default')
    })

    it('should fetch resources without namespace when namespace is "all"', async () => {
      const mockData = [{ name: 'pod1' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchResources(mockFetcher, 'Pod', 'all', null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/resources?kind=Pod')
    })

    it('should include cluster parameter when provided', async () => {
      const mockData = [{ name: 'pod1' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchResources(mockFetcher, 'Pod', 'default', 'cluster1')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/resources?kind=Pod&namespace=default&cluster=cluster1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchResources(mockFetcher, 'Pod', 'default', null)).rejects.toThrow('Failed to fetch Pod')
    })
  })

  describe('fetchClusterOverview', () => {
    it('should fetch cluster overview successfully', async () => {
      const mockData = { nodes: 3, pods: 10 }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchClusterOverview(mockFetcher)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/overview')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchClusterOverview(mockFetcher)).rejects.toThrow('Failed to fetch cluster overview')
    })
  })

  describe('fetchPrometheusStatus', () => {
    it('should fetch Prometheus status without cluster', async () => {
      const mockData = { connected: true }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchPrometheusStatus(mockFetcher, null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/prometheus/status?')
    })

    it('should fetch Prometheus status with cluster', async () => {
      const mockData = { connected: true }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchPrometheusStatus(mockFetcher, 'cluster1')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/prometheus/status?cluster=cluster1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchPrometheusStatus(mockFetcher, null)).rejects.toThrow('Failed to fetch Prometheus status')
    })
  })

  describe('fetchClusterMetrics', () => {
    it('should fetch cluster metrics without cluster', async () => {
      const mockData = { cpu: '50%', memory: '60%' }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchClusterMetrics(mockFetcher, null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/prometheus/cluster-overview?')
    })

    it('should fetch cluster metrics with cluster', async () => {
      const mockData = { cpu: '50%', memory: '60%' }
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchClusterMetrics(mockFetcher, 'cluster1')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/prometheus/cluster-overview?cluster=cluster1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchClusterMetrics(mockFetcher, null)).rejects.toThrow('Failed to fetch cluster metrics')
    })
  })

  describe('fetchNamespaces', () => {
    it('should fetch namespaces without cluster', async () => {
      const mockData = [{ name: 'default' }, { name: 'kube-system' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchNamespaces(mockFetcher, null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/namespaces?')
    })

    it('should fetch namespaces with cluster', async () => {
      const mockData = [{ name: 'default' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchNamespaces(mockFetcher, 'cluster1')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/namespaces?cluster=cluster1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchNamespaces(mockFetcher, null)).rejects.toThrow('Failed to fetch namespaces')
    })
  })

  describe('fetchHelmReleases', () => {
    it('should fetch Helm releases without cluster', async () => {
      const mockData = [{ name: 'release1', namespace: 'default' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchHelmReleases(mockFetcher, null)
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/helm/releases?')
    })

    it('should fetch Helm releases with cluster', async () => {
      const mockData = [{ name: 'release1' }]
      mockFetcher.mockResolvedValue({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchHelmReleases(mockFetcher, 'cluster1')
      expect(result).toEqual(mockData)
      expect(mockFetcher).toHaveBeenCalledWith('/api/helm/releases?cluster=cluster1')
    })

    it('should throw error when response is not ok', async () => {
      mockFetcher.mockResolvedValue({
        ok: false
      })

      await expect(fetchHelmReleases(mockFetcher, null)).rejects.toThrow('Failed to fetch helm releases')
    })
  })
})

