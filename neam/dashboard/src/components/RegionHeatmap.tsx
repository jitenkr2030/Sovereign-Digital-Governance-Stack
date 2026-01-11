import React, { useState, useMemo } from 'react';
import * as d3 from 'd3';
import type { RegionData, RegionStatus } from '../types';

interface RegionHeatmapProps {
  regions: RegionData[];
}

export const RegionHeatmap: React.FC<RegionHeatmapProps> = ({ regions }) => {
  const [selectedMetric, setSelectedMetric] = useState<string>('stressIndex');
  const [hoveredRegion, setHoveredRegion] = useState<RegionData | null>(null);

  const getStatusColor = (status: RegionStatus): string => {
    switch (status) {
      case 'healthy': return '#10b981';
      case 'caution': return '#84cc16';
      case 'warning': return '#f59e0b';
      case 'critical': return '#ef4444';
      default: return '#6b7280';
    }
  };

  const getStatusBgColor = (status: RegionStatus): string => {
    switch (status) {
      case 'healthy': return 'bg-green-100';
      case 'caution': return 'bg-lime-100';
      case 'warning': return 'bg-yellow-100';
      case 'critical': return 'bg-red-100';
      default: return 'bg-gray-100';
    }
  };

  // Calculate grid layout for heatmap
  const gridData = useMemo(() => {
    const cols = Math.ceil(Math.sqrt(regions.length));
    return regions.map((region, index) => ({
      ...region,
      col: index % cols,
      row: Math.floor(index / cols),
    }));
  }, [regions]);

  // Calculate color intensity based on value
  const getIntensityColor = (value: number, status: RegionStatus): string => {
    const baseColor = getStatusColor(status);
    const intensity = Math.min(value / 100, 1);
    
    // Parse hex to RGB and interpolate
    const hex = baseColor.replace('#', '');
    const r = parseInt(hex.substring(0, 2), 16);
    const g = parseInt(hex.substring(2, 4), 16);
    const b = parseInt(hex.substring(4, 6), 16);
    
    // Interpolate with white based on intensity
    const whiteRatio = 1 - intensity;
    const finalR = Math.round(r * intensity + 255 * whiteRatio);
    const finalG = Math.round(g * intensity + 255 * whiteRatio);
    const finalB = Math.round(b * intensity + 255 * whiteRatio);
    
    return `rgb(${finalR}, ${finalG}, ${finalB})`;
  };

  const formatMetricValue = (region: RegionData, metric: string): string => {
    switch (metric) {
      case 'stressIndex':
        return `${region.stressIndex.toFixed(1)}`;
      case 'inflationRate':
        return `${region.metrics.inflationRate.toFixed(2)}%`;
      case 'unemploymentRate':
        return `${region.metrics.unemploymentRate.toFixed(2)}%`;
      case 'gdpGrowth':
        return `${region.metrics.gdpGrowth >= 0 ? '+' : ''}${region.metrics.gdpGrowth.toFixed(2)}%`;
      default:
        return region.stressIndex.toFixed(1);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      {/* Metric Selector */}
      <div className="flex justify-end mb-4">
        <select
          value={selectedMetric}
          onChange={(e) => setSelectedMetric(e.target.value)}
          className="border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="stressIndex">Stress Index</option>
          <option value="inflationRate">Inflation Rate</option>
          <option value="unemploymentRate">Unemployment Rate</option>
          <option value="gdpGrowth">GDP Growth</option>
        </select>
      </div>

      {/* Heatmap Grid */}
      <div className="grid gap-2" style={{
        gridTemplateColumns: `repeat(auto-fill, minmax(100px, 1fr))`,
      }}>
        {gridData.map((region) => (
          <div
            key={region.id}
            className={`
              relative p-3 rounded-lg border-2 cursor-pointer transition-all hover:scale-105 hover:shadow-md
              ${getStatusBgColor(region.status)}
            `}
            style={{ borderColor: getStatusColor(region.status) }}
            onMouseEnter={() => setHoveredRegion(region)}
            onMouseLeave={() => setHoveredRegion(null)}
          >
            {/* Region Name */}
            <div className="text-xs font-semibold text-gray-800 truncate mb-1">
              {region.name}
            </div>
            
            {/* Metric Value */}
            <div className="text-lg font-bold text-gray-900">
              {formatMetricValue(region, selectedMetric)}
            </div>
            
            {/* Status Indicator */}
            <div className="absolute top-2 right-2">
              <div 
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: getStatusColor(region.status) }}
              />
            </div>
            
            {/* Alert/Intervention Count */}
            <div className="flex items-center justify-between mt-2 text-xs text-gray-600">
              {region.activeAlerts > 0 && (
                <span className="flex items-center text-orange-600">
                  <span className="w-4 h-4 rounded-full bg-orange-200 flex items-center justify-center text-[10px] font-medium mr-1">
                    {region.activeAlerts}
                  </span>
                  Alerts
                </span>
              )}
              {region.activeInterventions > 0 && (
                <span className="flex items-center text-blue-600">
                  <span className="w-4 h-4 rounded-full bg-blue-200 flex items-center justify-center text-[10px] font-medium mr-1">
                    {region.activeInterventions}
                  </span>
                  Active
                </span>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Tooltip */}
      {hoveredRegion && (
        <div 
          className="fixed z-50 bg-gray-900 text-white p-4 rounded-lg shadow-xl pointer-events-none"
          style={{
            left: '50%',
            bottom: '20px',
            transform: 'translateX(-50%)',
            maxWidth: '400px',
          }}
        >
          <div className="flex items-start justify-between mb-2">
            <div>
              <h4 className="font-semibold">{hoveredRegion.name}</h4>
              <p className="text-sm text-gray-400">{hoveredRegion.state}</p>
            </div>
            <span className={`px-2 py-1 rounded text-xs font-medium ${
              hoveredRegion.status === 'critical' ? 'bg-red-600' :
              hoveredRegion.status === 'warning' ? 'bg-yellow-600' :
              hoveredRegion.status === 'caution' ? 'bg-lime-600' : 'bg-green-600'
            }`}>
              {hoveredRegion.status.toUpperCase()}
            </span>
          </div>
          
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>
              <p className="text-gray-400">Stress Index</p>
              <p className="font-semibold">{hoveredRegion.stressIndex.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-gray-400">Inflation</p>
              <p className="font-semibold">{hoveredRegion.metrics.inflationRate.toFixed(2)}%</p>
            </div>
            <div>
              <p className="text-gray-400">Unemployment</p>
              <p className="font-semibold">{hoveredRegion.metrics.unemploymentRate.toFixed(2)}%</p>
            </div>
            <div>
              <p className="text-gray-400">GDP Growth</p>
              <p className={`font-semibold ${
                hoveredRegion.metrics.gdpGrowth >= 0 ? 'text-green-400' : 'text-red-400'
              }`}>
                {hoveredRegion.metrics.gdpGrowth >= 0 ? '+' : ''}{hoveredRegion.metrics.gdpGrowth.toFixed(2)}%
              </p>
            </div>
          </div>
          
          <div className="flex items-center justify-between mt-3 pt-2 border-t border-gray-700">
            <span className="text-xs text-gray-400">
              {hoveredRegion.activeAlerts} active alerts
            </span>
            <span className="text-xs text-gray-400">
              {hoveredRegion.activeInterventions} interventions
            </span>
          </div>
        </div>
      )}

      {/* Legend */}
      <div className="flex items-center justify-center gap-6 mt-4 pt-4 border-t">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded bg-green-500"></div>
          <span className="text-sm text-gray-600">Healthy (0-25)</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded bg-lime-500"></div>
          <span className="text-sm text-gray-600">Caution (26-50)</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded bg-yellow-500"></div>
          <span className="text-sm text-gray-600">Warning (51-75)</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded bg-red-500"></div>
          <span className="text-sm text-gray-600">Critical (76-100)</span>
        </div>
      </div>
    </div>
  );
};
