import { useRef, useEffect, useCallback } from 'react';
import * as d3 from 'd3';

interface UseD3Options {
  width?: number;
  height?: number;
  responsive?: boolean;
}

export function useD3<T extends SVGSVGElement>(
  renderChart: (svg: d3.Selection<T, unknown, null, undefined>, dimensions: { width: number; height: number }) => void | (() => void),
  dependencies: unknown[],
  options: UseD3Options = {}
) {
  const { width = 800, height = 400, responsive = true } = options;
  const svgRef = useRef<T>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const svg = d3.select(svgRef.current);
    
    // Clear previous content
    svg.selectAll('*').remove();

    // Get dimensions
    let chartWidth = width;
    let chartHeight = height;

    if (responsive && containerRef.current) {
      const containerWidth = containerRef.current.clientWidth;
      chartWidth = containerWidth || width;
    }

    // Create chart group
    const chartGroup = svg
      .attr('width', chartWidth)
      .attr('height', chartHeight)
      .attr('viewBox', `0 0 ${chartWidth} ${chartHeight}`)
      .attr('preserveAspectRatio', 'xMidYMid meet')
      .append('g');

    // Render chart
    cleanupRef.current = renderChart(chartGroup as d3.Selection<T, unknown, null, undefined>, {
      width: chartWidth,
      height: chartHeight,
    });

    return () => {
      if (cleanupRef.current) {
        cleanupRef.current();
      }
      svg.selectAll('*').remove();
    };
  }, [renderChart, width, height, responsive, ...dependencies]);

  // Handle resize
  useEffect(() => {
    if (!responsive || !containerRef.current) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width: newWidth } = entry.contentRect;
        if (newWidth > 0 && svgRef.current) {
          d3.select(svgRef.current)
            .attr('width', newWidth)
            .attr('viewBox', `0 0 ${newWidth} ${height}`);
        }
      }
    });

    resizeObserver.observe(containerRef.current);
    return () => resizeObserver.disconnect();
  }, [responsive, height]);

  return { svgRef, containerRef };
}

// Utility functions for D3 charts
export const d3Utils = {
  // Create scales
  createTimeScale(domain: [Date, Date], range: [number, number]) {
    return d3.scaleTime().domain(domain).range(range);
  },

  createLinearScale(domain: [number, number], range: [number, number]) {
    return d3.scaleLinear().domain(domain).range(range);
  },

  createBandScale(domain: string[], range: [number, number], padding = 0.1) {
    return d3.scaleBand().domain(domain).range(range).padding(padding);
  },

  createColorScale(domain: string[], colors: string[]) {
    return d3.scaleOrdinal<string>().domain(domain).range(colors);
  },

  // Create axes
  createAxisBottom(scale: d3.ScaleTime<number, number> | d3.ScaleLinear<number, number>) {
    return d3.axisBottom(scale).ticks(6).tickSizeOuter(0);
  },

  createAxisLeft(scale: d3.ScaleLinear<number, number>) {
    return d3.axisLeft(scale).ticks(5).tickSizeOuter(0);
  },

  // Line generator
  createLineGenerator() {
    return d3.line<{ date: Date; value: number }>()
      .x((d) => d.date.getTime())
      .y((d) => d.value)
      .curve(d3.curveMonotoneX);
  },

  // Area generator
  createAreaGenerator() {
    return d3.area<{ date: Date; value: number }>()
      .x((d) => d.date.getTime())
      .y0((d) => d.value)
      .y1((d) => d.value)
      .curve(d3.curveMonotoneX);
  },

  // Animation utilities
  transition(svg: d3.Selection<SVGGElement, unknown, null, undefined>, duration = 750) {
    return svg.transition().duration(duration);
  },

  // Tooltip utilities
  createTooltip(container: d3.Selection<SVGGElement, unknown, null, undefined>) {
    const tooltip = container
      .append('g')
      .attr('class', 'd3-tooltip-container')
      .style('display', 'none');

    tooltip
      .append('rect')
      .attr('class', 'tooltip-bg')
      .attr('rx', 4)
      .attr('fill', 'rgba(0,0,0,0.8)')
      .attr('stroke', 'rgba(255,255,255,0.2)');

    tooltip
      .append('text')
      .attr('class', 'tooltip-text')
      .attr('fill', 'white')
      .attr('font-size', '12px')
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em');

    return tooltip;
  },

  showTooltip(tooltip: d3.Selection<SVGGElement, unknown, null, undefined>, x: number, y: number, content: string) {
    const tooltipBg = tooltip.select<SVGRectElement>('.tooltip-bg');
    const tooltipText = tooltip.select<SVGTextElement>('.tooltip-text');

    tooltipText.text(content);
    const bbox = (tooltipText.node() as SVGTextElement).getBBox();
    
    tooltipBg
      .attr('x', x - bbox.width / 2 - 10)
      .attr('y', y - bbox.height / 2 - 6)
      .attr('width', bbox.width + 20)
      .attr('height', bbox.height + 12);

    tooltipText
      .attr('x', x)
      .attr('y', y);

    tooltip.style('display', null);
  },

  hideTooltip(tooltip: d3.Selection<SVGGElement, unknown, null, undefined>) {
    tooltip.style('display', 'none');
  },

  // Grid utilities
  addGridLines(
    svg: d3.Selection<SVGGElement, unknown, null, undefined>,
    width: number,
    height: number,
    scale: d3.ScaleLinear<number, number>,
    orientation: 'horizontal' | 'vertical' = 'horizontal'
  ) {
    const gridGroup = svg.append('g').attr('class', 'grid');

    if (orientation === 'horizontal') {
      const ticks = scale.ticks(5);
      gridGroup
        .selectAll('line')
        .data(ticks)
        .join('line')
        .attr('x1', 0)
        .attr('x2', width)
        .attr('y1', (d) => scale(d))
        .attr('y2', (d) => scale(d))
        .attr('stroke', 'currentColor')
        .attr('stroke-opacity', 0.1);
    } else {
      const ticks = scale.ticks(6);
      gridGroup
        .selectAll('line')
        .data(ticks)
        .join('line')
        .attr('y1', 0)
        .attr('y2', height)
        .attr('x1', (d) => scale(d))
        .attr('x2', (d) => scale(d))
        .attr('stroke', 'currentColor')
        .attr('stroke-opacity', 0.1);
    }
  },

  // Legend utilities
  createLegend(
    svg: d3.Selection<SVGGElement, unknown, null, undefined>,
    data: { label: string; color: string }[],
    x: number,
    y: number
  ) {
    const legend = svg.append('g').attr('class', 'legend');

    legend
      .selectAll('g')
      .data(data)
      .join('g')
      .attr('transform', (_, i) => `translate(${x}, ${y + i * 20})`)
      .each(function (d) {
        const g = d3.select(this);
        g.append('rect')
          .attr('width', 12)
          .attr('height', 12)
          .attr('rx', 2)
          .attr('fill', d.color);
        g.append('text')
          .attr('x', 18)
          .attr('y', 10)
          .attr('fill', 'currentColor')
          .attr('font-size', '12px')
          .text(d.label);
      });
  },

  // Responsive utility
  getResponsiveWidth(containerRef: React.RefObject<HTMLElement>, defaultWidth = 600): number {
    if (!containerRef.current) return defaultWidth;
    return containerRef.current.clientWidth || defaultWidth;
  },
};

export default d3Utils;
