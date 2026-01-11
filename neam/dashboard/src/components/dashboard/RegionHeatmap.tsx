/**
 * NEAM Dashboard - RegionHeatmap Component
 * Interactive choropleth map for regional economic visualization
 */

import React, { useState, useMemo, useCallback } from 'react';
import { ComposableMap, Geographies, Geography, ZoomableGroup } from 'react-simple-maps';
import { Tooltip } from 'react-tooltip';
import { MapPin, Info, Search } from 'lucide-react';
import type { HeatmapCell, RegionData, ColorScale } from '../../types';
import clsx from 'clsx';

interface RegionHeatmapProps {
  data: HeatmapCell[];
  regions?: RegionData[];
  config?: {
    metric?: string;
    colorScale?: ColorScale;
    showLegend?: boolean;
    interactive?: boolean;
  };
  onRegionClick?: (region: HeatmapCell) => void;
  onRegionHover?: (region: HeatmapCell | null) => void;
  className?: string;
  height?: number;
}

const DEFAULT_COLOR_SCALE: ColorScale = {
  low: '#10B981',    // Emerald 500 - Good
  medium: '#F59E0B', // Amber 500 - Caution
  high: '#EF4444',   // Red 500 - Warning
  critical: '#7C3AED', // Purple 600 - Critical
};

const INDIA_GEO_URL = 'https://raw.githubusercontent.com/deldersveld/topojson/master/maps/india.json';

// Utility function to get color based on value
const getColorByValue = (
  value: number,
  minValue: number,
  maxValue: number,
  colorScale: ColorScale
): string => {
  const normalized = (value - minValue) / (maxValue - minValue || 1);
  
  if (normalized < 0.25) return colorScale.low;
  if (normalized < 0.5) return colorScale.low;
  if (normalized < 0.75) return colorScale.medium;
  if (normalized < 0.9) return colorScale.high;
  return colorScale.critical;
};

export const RegionHeatmap: React.FC<RegionHeatmapProps> = ({
  data,
  regions = [],
  config = {},
  onRegionClick,
  onRegionHover,
  className,
  height = 500,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [hoveredRegion, setHoveredRegion] = useState<string | null>(null);
  const [selectedRegion, setSelectedRegion] = useState<string | null>(null);

  const {
    metric = 'stressIndex',
    colorScale = DEFAULT_COLOR_SCALE,
    showLegend = true,
    interactive = true,
  } = config;

  // Calculate min/max values for color scaling
  const { minValue, maxValue } = useMemo(() => {
    if (data.length === 0) return { minValue: 0, maxValue: 100 };
    const values = data.map((d) => d.value);
    return {
      minValue: Math.min(...values),
      maxValue: Math.max(...values),
    };
  }, [data]);

  // Filter data based on search
  const filteredData = useMemo(() => {
    if (!searchTerm) return data;
    return data.filter(
      (d) =>
        d.regionName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        d.regionId.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [data, searchTerm]);

  // Create a map of region data for quick lookup
  const regionDataMap = useMemo(() => {
    const map = new Map<string, HeatmapCell>();
    data.forEach((d) => map.set(d.regionId, d));
    return map;
  }, [data]);

  const handleGeographyClick = useCallback(
    (geo: any) => {
      const regionId = geo.properties.st_nm || geo.properties.name;
      const regionData = regionDataMap.get(regionId);
      
      if (regionData && interactive && onRegionClick) {
        setSelectedRegion(regionId);
        onRegionClick(regionData);
      }
    },
    [interactive, onRegionClick, regionDataMap]
  );

  const handleGeographyHover = useCallback(
    (geo: any) => {
      const regionId = geo.properties.st_nm || geo.properties.name;
      const regionData = regionDataMap.get(regionId);
      
      if (interactive && onRegionHover) {
        setHoveredRegion(regionId);
        onRegionHover(regionData || null);
      }
    },
    [interactive, onRegionHover, regionDataMap]
  );

  const handleGeographyLeave = useCallback(() => {
    if (interactive && onRegionHover) {
      setHoveredRegion(null);
      onRegionHover(null);
    }
  }, [interactive, onRegionHover]);

  return (
    <div className={clsx('bg-white rounded-xl border border-slate-200 p-4', className)}>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <MapPin className="w-5 h-5 text-slate-500" />
          <h3 className="text-lg font-semibold text-slate-800">Regional Economic Heatmap</h3>
          <Info className="w-4 h-4 text-slate-400" data-tooltip-id="heatmap-tooltip" />
        </div>
        
        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Search region..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9 pr-4 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Legend */}
      {showLegend && (
        <div className="flex items-center gap-4 mb-4">
          <div className="flex items-center gap-2">
            <span className="text-xs text-slate-500">Performance:</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: colorScale.low }} />
            <span className="text-xs text-slate-600">Good</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: colorScale.medium }} />
            <span className="text-xs text-slate-600">Caution</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: colorScale.high }} />
            <span className="text-xs text-slate-600">Warning</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: colorScale.critical }} />
            <span className="text-xs text-slate-600">Critical</span>
          </div>
        </div>
      )}

      {/* Map */}
      <div className="relative" style={{ height }}>
        <ComposableMap
          projection="geoMercator"
          projectionConfig={{
            scale: 350,
            center: [78.9629, 20.5937], // India center
          }}
          style={{ width: '100%', height: '100%' }}
        >
          <ZoomableGroup center={[78.9629, 20.5937]} zoom={1}>
            <Geographies geography={INDIA_GEO_URL}>
              {({ geographies }) =>
                geographies.map((geo) => {
                  const regionName = geo.properties.st_nm || geo.properties.name;
                  const regionData = regionDataMap.get(regionName);
                  const isHovered = hoveredRegion === regionName;
                  const isSelected = selectedRegion === regionName;
                  const value = regionData?.value ?? 0;
                  const fillColor = getColorByValue(value, minValue, maxValue, colorScale);

                  return (
                    <Geography
                      key={geo.rsmKey}
                      geography={geo}
                      onClick={() => handleGeographyClick(geo)}
                      onMouseEnter={() => handleGeographyHover(geo)}
                      onMouseLeave={handleGeographyLeave}
                      style={{
                        default: {
                          fill: fillColor,
                          stroke: isSelected ? '#3B82F6' : isHovered ? '#64748B' : '#CBD5E1',
                          strokeWidth: isSelected ? 2 : isHovered ? 1.5 : 0.5,
                          outline: 'none',
                        },
                        hover: {
                          fill: fillColor,
                          stroke: '#3B82F6',
                          strokeWidth: 2,
                          outline: 'none',
                          opacity: 0.8,
                        },
                        pressed: {
                          fill: fillColor,
                          stroke: '#2563EB',
                          strokeWidth: 2,
                          outline: 'none',
                        },
                      }}
                    />
                  );
                })
              }
            </Geographies>
          </ZoomableGroup>
        </ComposableMap>

        {/* Tooltip */}
        {hoveredRegion && (
          <div className="absolute bottom-4 left-4 bg-slate-800 text-white px-3 py-2 rounded-lg text-sm shadow-lg">
            <p className="font-medium">{hoveredRegion}</p>
            <p className="text-slate-300">
              {metric}: {regionDataMap.get(hoveredRegion)?.value.toFixed(2) || 'N/A'}
            </p>
          </div>
        )}
      </div>

      {/* Region List */}
      {searchTerm && (
        <div className="mt-4 border-t border-slate-200 pt-4">
          <p className="text-sm text-slate-500 mb-2">
            Matching regions ({filteredData.length}):
          </p>
          <div className="flex flex-wrap gap-2">
            {filteredData.slice(0, 10).map((region) => (
              <button
                key={region.regionId}
                onClick={() => handleGeographyClick({ properties: { st_nm: region.regionId } } as any)}
                className={clsx(
                  'px-3 py-1 rounded-full text-xs font-medium transition-colors',
                  'hover:bg-slate-100',
                  selectedRegion === region.regionId
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-slate-100 text-slate-600'
                )}
              >
                {region.regionName}
              </button>
            ))}
          </div>
        </div>
      )}

      <Tooltip id="heatmap-tooltip" content="Click on a region to drill down into detailed analytics" />
    </div>
  );
};

// Mini heatmap for sidebar/small spaces
export const MiniHeatmap: React.FC<{
  data: HeatmapCell[];
  maxItems?: number;
  onItemClick?: (region: HeatmapCell) => void;
}> = ({ data, maxItems = 5, onItemClick }) => {
  const sortedData = useMemo(() => {
    return [...data]
      .sort((a, b) => b.value - a.value)
      .slice(0, maxItems);
  }, [data, maxItems]);

  return (
    <div className="space-y-2">
      {sortedData.map((region) => (
        <div
          key={region.regionId}
          onClick={() => onItemClick?.(region)}
          className="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-50 cursor-pointer transition-colors"
        >
          <div
            className="w-3 h-3 rounded-full"
            style={{
              backgroundColor: getColorByValue(
                region.value,
                Math.min(...data.map((d) => d.value)),
                Math.max(...data.map((d) => d.value)),
                DEFAULT_COLOR_SCALE
              ),
            }}
          />
          <span className="flex-1 text-sm text-slate-600 truncate">
            {region.regionName}
          </span>
          <span className="text-sm font-medium text-slate-800">
            {region.value.toFixed(1)}
          </span>
        </div>
      ))}
    </div>
  );
};

export default RegionHeatmap;
