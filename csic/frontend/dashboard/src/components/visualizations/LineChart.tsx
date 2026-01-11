import { useD3 } from '../../hooks/useD3';
import * as d3 from 'd3';

interface LineChartProps {
  data: Array<{
    timestamp: string;
    value: number;
    category?: string;
  }>;
  width?: number;
  height?: number;
  categories?: string[];
  colors?: string[];
  showGrid?: boolean;
  showTooltip?: boolean;
  showLegend?: boolean;
  animate?: boolean;
  yAxisLabel?: string;
  xAxisFormat?: 'time' | 'date' | 'number';
  areaFill?: boolean;
}

export default function LineChart({
  data,
  width = 800,
  height = 400,
  categories = [],
  colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'],
  showGrid = true,
  showTooltip = true,
  showLegend = true,
  animate = true,
  yAxisLabel,
  xAxisFormat = 'time',
  areaFill = false,
}: LineChartProps) {
  const renderChart = (svg: d3.Selection<SVGSVGElement, unknown, null, undefined>, dimensions: { width: number; height: number }) => {
    const { width: w, height: h } = dimensions;
    const margin = { top: 20, right: 30, bottom: 40, left: 60 };
    const innerWidth = w - margin.left - margin.right;
    const innerHeight = h - margin.top - margin.bottom;

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    // Parse dates if needed
    const parsedData = data.map((d) => ({
      ...d,
      date: new Date(d.timestamp),
    }));

    // Get unique categories
    const uniqueCategories = categories.length > 0 ? categories : [...new Set(data.map((d) => d.category || 'default'))];
    const colorScale = d3.scaleOrdinal<string>().domain(uniqueCategories).range(colors);

    // Create scales
    const xScale = d3
      .scaleTime()
      .domain(d3.extent(parsedData, (d) => d.date) as [Date, Date])
      .range([0, innerWidth]);

    const yScale = d3
      .scaleLinear()
      .domain([0, d3.max(parsedData, (d) => d.value) as number])
      .nice()
      .range([innerHeight, 0]);

    // Add grid lines
    if (showGrid) {
      g.append('g')
        .attr('class', 'grid')
        .selectAll('line')
        .data(yScale.ticks(5))
        .join('line')
        .attr('x1', 0)
        .attr('x2', innerWidth)
        .attr('y1', (d) => yScale(d))
        .attr('y2', (d) => yScale(d))
        .attr('stroke', 'currentColor')
        .attr('stroke-opacity', 0.1);
    }

    // Create line generator
    const line = d3
      .line<typeof parsedData[0]>()
      .x((d) => xScale(d.date))
      .y((d) => yScale(d.value))
      .curve(d3.curveMonotoneX);

    // Create area generator if needed
    const area = areaFill
      ? d3
          .area<typeof parsedData[0]>()
          .x((d) => xScale(d.date))
          .y0(innerHeight)
          .y1((d) => yScale(d.value))
          .curve(d3.curveMonotoneX)
      : null;

    // Draw areas and lines for each category
    uniqueCategories.forEach((category) => {
      const categoryData = parsedData.filter((d) => (d.category || 'default') === category);
      const color = colorScale(category);

      if (areaFill && area) {
        const areaPath = g
          .append('path')
          .datum(categoryData)
          .attr('fill', color)
          .attr('fill-opacity', 0.1)
          .attr('d', area);

        if (animate) {
          areaPath
            .attr('opacity', 0)
            .transition()
            .duration(1000)
            .attr('opacity', 1);
        }
      }

      const linePath = g
        .append('path')
        .datum(categoryData)
        .attr('fill', 'none')
        .attr('stroke', color)
        .attr('stroke-width', 2)
        .attr('d', line);

      if (animate) {
        const totalLength = linePath.node()?.getTotalLength() || 0;
        linePath
          .attr('stroke-dasharray', `${totalLength} ${totalLength}`)
          .attr('stroke-dashoffset', totalLength)
          .transition()
          .duration(1500)
          .ease(d3.easeLinear)
          .attr('stroke-dashoffset', 0);
      }

      // Add dots
      g.selectAll(`.dot-${category}`)
        .data(categoryData)
        .join('circle')
        .attr('class', `dot-${category}`)
        .attr('cx', (d) => xScale(d.date))
        .attr('cy', (d) => yScale(d.value))
        .attr('r', 4)
        .attr('fill', color)
        .attr('stroke', 'white')
        .attr('stroke-width', 2)
        .style('cursor', 'pointer')
        .style('opacity', 0);
    });

    // Add axes
    const xAxis = d3.axisBottom(xScale).ticks(6).tickSizeOuter(0);
    const yAxis = d3.axisLeft(yScale).ticks(5).tickSizeOuter(0);

    g.append('g')
      .attr('transform', `translate(0,${innerHeight})`)
      .call(xAxis)
      .attr('class', 'axis')
      .selectAll('text')
      .style('font-size', '11px');

    g.append('g')
      .call(yAxis)
      .attr('class', 'axis')
      .selectAll('text')
      .style('font-size', '11px');

    // Add y-axis label
    if (yAxisLabel) {
      g.append('text')
        .attr('transform', 'rotate(-90)')
        .attr('y', -margin.left + 15)
        .attr('x', -innerHeight / 2)
        .attr('text-anchor', 'middle')
        .attr('fill', 'currentColor')
        .attr('font-size', '12px')
        .text(yAxisLabel);
    }

    // Add tooltip
    if (showTooltip) {
      const tooltip = g.append('g').attr('class', 'tooltip').style('display', 'none');

      tooltip
        .append('rect')
        .attr('width', 120)
        .attr('height', 50)
        .attr('rx', 4)
        .attr('fill', 'rgba(0,0,0,0.8)')
        .attr('stroke', 'rgba(255,255,255,0.2)');

      const tooltipText = tooltip.append('text').attr('fill', 'white').attr('font-size', '12px');

      const overlay = g
        .append('rect')
        .attr('width', innerWidth)
        .attr('height', innerHeight)
        .attr('fill', 'transparent')
        .on('mousemove', function (event) {
          const [mx] = d3.pointer(event);
          const date = xScale.invert(mx);
          const bisect = d3.bisector((d: typeof parsedData[0]) => d.date).left;
          const idx = bisect(parsedData, date);
          const d0 = parsedData[idx - 1];
          const d1 = parsedData[idx];
          const d = d1 && date.getTime() - d0.date.getTime() > d1.date.getTime() - date.getTime() ? d1 : d0;

          if (d) {
            const x = xScale(d.date);
            const y = yScale(d.value);

            tooltip
              .attr('transform', `translate(${Math.min(x, innerWidth - 130)},${Math.max(y - 60, 10)})`)
              .style('display', null);

            tooltipText.selectAll('tspan').remove();
            tooltipText.append('tspan').attr('x', 10).attr('y', 20).text(d.date.toLocaleString());
            tooltipText.append('tspan').attr('x', 10).attr('y', 38).text(`Value: ${d.value.toLocaleString()}`);
          }
        })
        .on('mouseleave', () => {
          tooltip.style('display', 'none');
        });
    }

    // Add legend
    if (showLegend && uniqueCategories.length > 1) {
      const legend = svg
        .append('g')
        .attr('transform', `translate(${w - margin.right - 100}, ${margin.top})`);

      uniqueCategories.forEach((category, i) => {
        const legendItem = legend
          .append('g')
          .attr('transform', `translate(0, ${i * 20})`);

        legendItem
          .append('rect')
          .attr('width', 12)
          .attr('height', 12)
          .attr('rx', 2)
          .attr('fill', colorScale(category));

        legendItem
          .append('text')
          .attr('x', 18)
          .attr('y', 10)
          .attr('fill', 'currentColor')
          .attr('font-size', '11px')
          .text(category);
      });
    }

    return () => {
      // Cleanup function
    };
  };

  const { containerRef } = useD3(renderChart, [data, categories, colors, showGrid, showTooltip, showLegend], {
    width,
    height,
    responsive: true,
  });

  return (
    <div ref={containerRef} className="w-full">
      <svg ref={null} className="d3-chart" />
    </div>
  );
}
