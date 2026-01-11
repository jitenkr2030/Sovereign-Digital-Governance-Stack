import React, { useEffect, useRef, useState, useCallback } from 'react';
import * as d3 from 'd3';
import { Card } from '../../ui/Card';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import './RegionalMap.css';

export interface RegionData {
  id: string;
  name: string;
  value: number;
  unit: string;
  status: 'healthy' | 'warning' | 'critical';
  metrics?: Record<string, number>;
}

export interface MapConfig {
  projection: 'geoMercator' | 'geoOrthographic' | 'geoAlbers';
  center: [number, number];
  scale: number;
  zoom: { min: number; max: number };
}

export interface RegionalMapProps {
  /** GeoJSON data for regions */
  geoJson: GeoJSON.FeatureCollection;
  /** Economic data for each region */
  data: RegionData[];
  /** Map configuration */
  config?: Partial<MapConfig>;
  /** Color scale for values */
  colorScale?: 'single' | 'diverging' | 'threshold';
  /** Value range for color mapping */
  valueRange?: [number, number];
  /** Show tooltips */
  showTooltips?: boolean;
  /** Click handler for regions */
  onRegionClick?: (region: RegionData) => void;
  /** Hover handler for regions */
  onRegionHover?: (region: RegionData | null) => void;
  /** Is loading */
  isLoading?: boolean;
  /** Height of the map container */
  height?: number | string;
  /** Custom class name */
  className?: string;
}

/**
 * Default map configuration
 */
const defaultConfig: MapConfig = {
  projection: 'geoMercator',
  center: [-98, 39],
  scale: 400,
  zoom: { min: 1, max: 8 },
};

/**
 * Status colors for regions
 */
const statusColors = {
  healthy: '#10B981',
  warning: '#F59E0B',
  critical: '#EF4444',
};

/**
 * Color scale based on value
 */
const getColorScale = (value: number, range: [number, number]): string => {
  const [min, max] = range;
  const normalized = Math.max(0, Math.min(1, (value - min) / (max - min || 1)));
  
  // Gradient from healthy (green) to warning (yellow) to critical (red)
  if (normalized < 0.5) {
    const t = normalized * 2;
    return interpolateColor(statusColors.healthy, statusColors.warning, t);
  } else {
    const t = (normalized - 0.5) * 2;
    return interpolateColor(statusColors.warning, statusColors.critical, t);
  }
};

// Helper function to interpolate between colors
const interpolateColor = (color1: string, color2: string, factor: number): string => {
  const hex = (c: string) => parseInt(c.slice(1), 16);
  const r1 = (hex(color1) >> 16) & 255;
  const g1 = (hex(color1) >> 8) & 255;
  const b1 = hex(color1) & 255;
  const r2 = (hex(color2) >> 16) & 255;
  const g2 = (hex(color2) >> 8) & 255;
  const b2 = hex(color2) & 255;
  
  const r = Math.round(r1 + factor * (r2 - r1));
  const g = Math.round(g1 + factor * (g2 - g1));
  const b = Math.round(b1 + factor * (b2 - b1));
  
  return `#${((r << 16) | (g << 8) | b).toString(16).padStart(6, '0')}`;
};

/**
 * RegionalMap Component
 * 
 * Interactive choropleth map for visualizing regional economic data using D3.js.
 */
export const RegionalMap: React.FC<RegionalMapProps> = ({
  geoJson,
  data,
  config = {},
  colorScale = 'single',
  valueRange = [0, 100],
  showTooltips = true,
  onRegionClick,
  onRegionHover,
  isLoading = false,
  height = 400,
  className = '',
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: height });
  const [activeRegion, setActiveRegion] = useState<string | null>(null);

  // Merge config with defaults
  const mapConfig = { ...defaultConfig, ...config };

  // Handle resize
  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const { width } = containerRef.current.getBoundingClientRect();
        setDimensions({ width, height: typeof height === 'number' ? height : 400 });
      }
    };

    updateDimensions();
    window.addEventListener('resize', updateDimensions);
    return () => window.removeEventListener('resize', updateDimensions);
  }, [height]);

  // Draw map
  useEffect(() => {
    if (!svgRef.current || !geoJson || dimensions.width === 0) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    const { width, height } = dimensions;
    const margin = { top: 20, right: 20, bottom: 20, left: 20 };

    // Create projection
    const projection = d3.geoMercator()
      .center(mapConfig.center)
      .scale(mapConfig.scale * (width / 500))
      .translate([width / 2, height / 2]);

    const pathGenerator = d3.geoPath().projection(projection);

    // Create zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([mapConfig.zoom.min, mapConfig.zoom.max])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoom);

    const g = svg.append('g');

    // Draw regions
    const regions = g.selectAll('.region')
      .data(geoJson.features)
      .enter()
      .append('path')
      .attr('class', 'region')
      .attr('d', pathGenerator as d3.GeoPath<GeoJSON.Feature, d3.GeoPermissibleObjects>)
      .attr('fill', (d) => {
        const regionData = data.find(r => r.id === d.properties?.name || r.id === d.id);
        if (!regionData) return 'var(--color-bg-tertiary)';
        
        if (colorScale === 'threshold') {
          return statusColors[regionData.status];
        }
        return getColorScale(regionData.value, valueRange);
      })
      .attr('stroke', 'var(--border-default)')
      .attr('stroke-width', 0.5)
      .style('cursor', 'pointer')
      .style('transition', 'opacity 0.2s ease');

    // Add hover effects
    regions
      .on('mouseenter', function(event, d) {
        const regionData = data.find(r => r.id === d.properties?.name || r.id === d.id);
        
        d3.select(this)
          .attr('stroke', 'var(--color-primary)')
          .attr('stroke-width', 2);
        
        setActiveRegion(regionData?.id || null);
        onRegionHover?.(regionData || null);

        // Show tooltip
        if (showTooltips && tooltipRef.current && regionData) {
          const tooltip = tooltipRef.current;
          tooltip.style.opacity = '1';
          tooltip.style.left = `${event.pageX + 10}px`;
          tooltip.style.top = `${event.pageY - 10}px`;
          tooltip.innerHTML = `
            <div class="map-tooltip-title">${regionData.name}</div>
            <div class="map-tooltip-value">${regionData.value.toLocaleString()} ${regionData.unit}</div>
            <div class="map-tooltip-status status-${regionData.status}">${regionData.status}</div>
          `;
        }
      })
      .on('mouseleave', function() {
        d3.select(this)
          .attr('stroke', 'var(--border-default)')
          .attr('stroke-width', 0.5);
        
        setActiveRegion(null);
        onRegionHover?.(null);

        if (showTooltips && tooltipRef.current) {
          tooltipRef.current.style.opacity = '0';
        }
      })
      .on('click', function(event, d) {
        const regionData = data.find(r => r.id === d.properties?.name || r.id === d.id);
        if (regionData) {
          onRegionClick?.(regionData);
        }
      });

  }, [geoJson, data, dimensions, mapConfig, colorScale, valueRange, showTooltips, onRegionClick, onRegionHover]);

  if (isLoading) {
    return (
      <Card className={`regional-map ${className}`}>
        <div className="regional-map-loading">
          <LoadingSpinner size="lg" label="Loading map..." />
        </div>
      </Card>
    );
  }

  return (
    <div ref={containerRef} className={`regional-map-container ${className}`}>
      <Card className="regional-map">
        <svg
          ref={svgRef}
          width={dimensions.width}
          height={dimensions.height}
          className="regional-map-svg"
        />
        
        {/* Legend */}
        <div className="map-legend">
          <div className="map-legend-title">Economic Health</div>
          <div className="map-legend-scale">
            <div className="map-legend-gradient" />
            <div className="map-legend-labels">
              <span>{valueRange[0]}</span>
              <span>{valueRange[1]}</span>
            </div>
          </div>
          <div className="map-legend-items">
            <div className="map-legend-item">
              <span className="map-legend-color" style={{ background: statusColors.healthy }} />
              <span>Healthy</span>
            </div>
            <div className="map-legend-item">
              <span className="map-legend-color" style={{ background: statusColors.warning }} />
              <span>Warning</span>
            </div>
            <div className="map-legend-item">
              <span className="map-legend-color" style={{ background: statusColors.critical }} />
              <span>Critical</span>
            </div>
          </div>
        </div>

        {/* Tooltip */}
        {showTooltips && (
          <div ref={tooltipRef} className="map-tooltip" />
        )}

        {/* Region Info Panel */}
        {activeRegion && (
          <div className="map-region-info">
            {(() => {
              const region = data.find(r => r.id === activeRegion);
              if (!region) return null;
              return (
                <>
                  <h4>{region.name}</h4>
                  <p className="map-info-value">
                    {region.value.toLocaleString()} {region.unit}
                  </p>
                  <span className={`map-info-status status-${region.status}`}>
                    {region.status}
                  </span>
                </>
              );
            })()}
          </div>
        )}
      </Card>
    </div>
  );
};

export default RegionalMap;
