import api from './api';
import { EnergyMetrics, EnergyForecast, RegionalEnergyData, GridStatus } from '../types';

interface EnergyQueryParams {
  region?: string;
  startDate?: string;
  endDate?: string;
  interval?: 'minute' | 'hour' | 'day' | 'week' | 'month';
}

interface ForecastParams {
  region: string;
  horizon: '1h' | '6h' | '24h' | '7d' | '30d';
  startDate?: string;
}

export const energyService = {
  // Real-time Telemetry
  async getCurrentMetrics(region?: string): Promise<EnergyMetrics> {
    const params = region ? `?region=${region}` : '';
    return api.get<EnergyMetrics>(`/telemetry/current${params}`);
  },

  async getMetricsHistory(params: EnergyQueryParams): Promise<EnergyMetrics[]> {
    const queryParams = new URLSearchParams();
    if (params.region) queryParams.append('region', params.region);
    if (params.startDate) queryParams.append('startDate', params.startDate);
    if (params.endDate) queryParams.append('endDate', params.endDate);
    if (params.interval) queryParams.append('interval', params.interval);

    return api.get<EnergyMetrics[]>(`/telemetry/history?${queryParams}`);
  },

  async getMetricsStream(region: string, callback: (data: EnergyMetrics) => void): Promise<() => void> {
    const eventSource = new EventSource(`/api/telemetry/stream?region=${encodeURIComponent(region)}`);
    
    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      callback(data);
    };

    eventSource.onerror = () => {
      eventSource.close();
    };

    return () => eventSource.close();
  },

  // Forecasting
  async getLoadForecast(params: ForecastParams): Promise<EnergyForecast[]> {
    const queryParams = new URLSearchParams({
      region: params.region,
      horizon: params.horizon,
    });
    if (params.startDate) queryParams.append('startDate', params.startDate);

    return api.get<EnergyForecast[]>(`/forecast/load?${queryParams}`);
  },

  async getCarbonForecast(region: string, horizon: string): Promise<Array<{
    timestamp: string;
    predicted: number;
    confidenceLower: number;
    confidenceUpper: number;
  }>> {
    return api.get(`/forecast/carbon?region=${region}&horizon=${horizon}`);
  },

  // Regional Data
  async getRegionalData(): Promise<RegionalEnergyData[]> {
    return api.get<RegionalEnergyData[]>('/energy/regional');
  },

  async getRegionalConsumption(region: string, period?: string): Promise<Array<{
    timestamp: string;
    consumption: number;
    demand: number;
  }>> {
    const params = period ? `?period=${period}` : '';
    return api.get(`/energy/regional/${region}/consumption${params}`);
  },

  async getRegionalRenewableMix(region: string): Promise<{
    solar: number;
    wind: number;
    hydro: number;
    nuclear: number;
    fossil: number;
    other: number;
  }> {
    return api.get(`/energy/regional/${region}/renewable-mix`);
  },

  // Grid Status
  async getGridStatus(region?: string): Promise<GridStatus[]> {
    const params = region ? `?region=${region}` : '';
    return api.get<GridStatus[]>(`/grid/status${params}`);
  },

  async getGridAlerts(): Promise<Array<{
    id: string;
    severity: 'info' | 'warning' | 'critical';
    region: string;
    message: string;
    timestamp: string;
  }>> {
    return api.get('/grid/alerts');
  },

  // Carbon Emissions
  async getCarbonEmissions(params: EnergyQueryParams): Promise<Array<{
    timestamp: string;
    emissions: number;
    intensity: number;
  }>> {
    const queryParams = new URLSearchParams();
    if (params.region) queryParams.append('region', params.region);
    if (params.startDate) queryParams.append('startDate', params.startDate);
    if (params.endDate) queryParams.append('endDate', params.endDate);

    return api.get(`/energy/carbon/emissions?${queryParams}`);
  },

  async getCarbonIntensityMap(): Promise<Array<{
    region: string;
    intensity: number;
    lat: number;
    lng: number;
  }>> {
    return api.get('/energy/carbon/intensity-map');
  },

  // Statistics and Analytics
  async getEnergyStats(): Promise<{
    totalConsumption: number;
    peakDemand: number;
    averageLoad: number;
    renewablePercentage: number;
    carbonEmissions: number;
    efficiency: number;
  }> {
    return api.get('/energy/stats');
  },

  async getConsumptionTrend(days: number = 30): Promise<Array<{
    date: string;
    consumption: number;
    demand: number;
    renewable: number;
  }>> {
    return api.get(`/energy/trend?days=${days}`);
  },

  async getPeakDemandTimes(region?: string): Promise<Array<{
    timestamp: string;
    demand: number;
    region: string;
  }>> {
    const params = region ? `?region=${region}` : '';
    return api.get(`/energy/peak-demand${params}`);
  },

  async getEnergyEfficiencyMetrics(): Promise<{
    overall: number;
    byRegion: Array<{ region: string; efficiency: number }>;
    trend: Array<{ date: string; efficiency: number }>;
  }> {
    return api.get('/energy/efficiency');
  },

  // Load Balancing
  async getLoadBalancingSuggestions(): Promise<Array<{
    region: string;
    action: 'increase' | 'decrease' | 'maintain';
    amount: number;
    reason: string;
  }>> {
    return api.get('/grid/load-balancing/suggestions');
  },

  async getCapacityForecast(region: string, days: number = 30): Promise<Array<{
    date: string;
    capacity: number;
    demand: number;
    margin: number;
  }>> {
    return api.get(`/grid/capacity/${region}?days=${days}`);
  },

  // Historical Analysis
  async getHistoricalAnalysis(params: {
    metric: 'consumption' | 'demand' | 'emissions';
    region?: string;
    period: 'day' | 'week' | 'month' | 'year';
  }): Promise<{
    current: number;
    previous: number;
    change: number;
    data: Array<{ timestamp: string; value: number }>;
  }> {
    const queryParams = new URLSearchParams({
      metric: params.metric,
      period: params.period,
    });
    if (params.region) queryParams.append('region', params.region);

    return api.get(`/energy/historical?${queryParams}`);
  },
};

export default energyService;
