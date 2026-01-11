import { useD3 } from '../../hooks/useD3';
import * as d3 from 'd3';

interface BarChartProps {
  data: Array<{
    label: string;
    value: number;
    color?: string;
  }>;
  width?: number;
  height?: number;
  colors?: string[];
  showValues?: boolean;
  showGrid?: boolean;
  showTooltip?: boolean;
  horizontal?: boolean;
  animate?: boolean;
  barRadius?: number;
}

export default function BarChart({
  data,
  width = 800,
  height = 400,
  colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'],
  showValues = true,
  showGrid = true,
  showTooltip = true,
  horizontal = false,
  animate = true,
  barRadius = 4,
}: BarChartProps) {
  const renderChart = (svg: d3.Selection<SVGSVGElement, unknown, null, undefined>, dimensions: { width: number; height: number }) => {
    const { width: w, height: h } = dimensions;
    const margin = { top: 20, right: 30, bottom: 50, left: horizontal ? 120 : 60 };
    const innerWidth = w - margin.left - margin.right;
    const innerHeight = h - margin.top - margin.bottom;

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    const colorScale = d3.scaleOrdinal<string>().domain(data.map((d) => d.label)).range(colors);

    if (horizontal) {
      // Horizontal bar chart
      const xScale = d3
        .scaleLinear()
        .domain([0, d3.max(data, (d) => d.value) as number])
        .nice()
        .range([0, innerWidth]);

      const yScale = d3
        .scaleBand()
        .domain(data.map((d) => d.label))
        .range([0, innerHeight])
        .padding(0.3);

      // Grid lines
      if (showGrid) {
        g.append('g')
          .attr('class', 'grid')
          .selectAll('line')
          .data(xScale.ticks(5))
          .join('line')
          .attr('x1', (d) => xScale(d))
          .attr('x2', (d) => xScale(d))
          .attr('y1', 0)
          .attr('y2', innerHeight)
          .attr('stroke', 'currentColor')
          .attr('stroke-opacity', 0.1);
      }

      // Bars
      const bars = g
        .selectAll('.bar')
        .data(data)
        .join('rect')
        .attr('class', 'bar')
        .attr('y', (d) => yScale(d.label) as number)
        .attr('height', yScale.bandwidth())
        .attr('fill', (d) => d.color || colorScale(d.label))
        .attr('rx', barRadius);

      if (animate) {
        bars
          .attr('x', 0)
          .attr('width', 0)
          .transition()
          .duration(800)
          .ease(d3.easeElastic.amplitude(1).period(0.5))
          .attr('width', (d) => xScale(d.value));
      } else {
        bars.attr('x', 0).attr('width', (d) => xScale(d.value));
      }

      // Values
      if (showValues) {
        const values = g
          .selectAll('.value')
          .data(data)
          .join('text')
          .attr('class', 'value')
          .attr('y', (d) => (yScale(d.label) as number) + yScale.bandwidth() / 2)
          .attr('dy', '0.35em')
          .attr('fill', 'currentColor')
          .attr('font-size', '12px')
          .attr('font-weight', '500');

        if (animate) {
          values
            .attr('x', 5)
            .attr('opacity', 0)
            .transition()
            .delay(800)
            .duration(300)
            .attr('x', (d) => xScale(d.value) + 8)
            .attr('opacity', 1);
        } else {
          values.attr('x', (d) => xScale(d.value) + 8).attr('opacity', 1);
        }
      }

      // Y axis
      g.append('g')
        .call(d3.axisLeft(yScale).tickSize(0))
        .attr('class', 'axis')
        .selectAll('text')
        .style('font-size', '12px');

      // X axis
      g.append('g')
        .attr('transform', `translate(0,${innerHeight})`)
        .call(d3.axisBottom(xScale).ticks(5).tickSizeOuter(0))
        .attr('class', 'axis')
        .selectAll('text')
        .style('font-size', '11px');
    } else {
      // Vertical bar chart
      const xScale = d3
        .scaleBand()
        .domain(data.map((d) => d.label))
        .range([0, innerWidth])
        .padding(0.3);

      const yScale = d3
        .scaleLinear()
        .domain([0, d3.max(data, (d) => d.value) as number])
        .nice()
        .range([innerHeight, 0]);

      // Grid lines
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

      // Bars
      const bars = g
        .selectAll('.bar')
        .data(data)
        .join('rect')
        .attr('class', 'bar')
        .attr('x', (d) => xScale(d.label) as number)
        .attr('width', xScale.bandwidth())
        .attr('fill', (d) => d.color || colorScale(d.label))
        .attr('rx', barRadius);

      if (animate) {
        bars
          .attr('y', innerHeight)
          .attr('height', 0)
          .transition()
          .duration(800)
          .ease(d3.easeElastic.amplitude(1).period(0.5))
          .attr('y', (d) => yScale(d.value))
          .attr('height', (d) => innerHeight - yScale(d.value));
      } else {
        bars
          .attr('y', (d) => yScale(d.value))
          .attr('height', (d) => innerHeight - yScale(d.value));
      }

      // Values
      if (showValues) {
        const values = g
          .selectAll('.value')
          .data(data)
          .join('text')
          .attr('class', 'value')
          .attr('x', (d) => (xScale(d.label) as number) + xScale.bandwidth() / 2)
          .attr('text-anchor', 'middle')
          .attr('fill', 'currentColor')
          .attr('font-size', '12px')
          .attr('font-weight', '500');

        if (animate) {
          values
            .attr('y', innerHeight)
            .attr('opacity', 0)
            .transition()
            .delay(800)
            .duration(300)
            .attr('y', (d) => yScale(d.value) - 8)
            .attr('opacity', 1);
        } else {
          values.attr('y', (d) => yScale(d.value) - 8).attr('opacity', 1);
        }
      }

      // X axis
      g.append('g')
        .attr('transform', `translate(0,${innerHeight})`)
        .call(d3.axisBottom(xScale).tickSize(0))
        .attr('class', 'axis')
        .selectAll('text')
        .style('font-size', '11px')
        .attr('transform', 'rotate(-45)')
        .attr('text-anchor', 'end');

      // Y axis
      g.append('g')
        .call(d3.axisLeft(yScale).ticks(5).tickSizeOuter(0))
        .attr('class', 'axis')
        .selectAll('text')
        .style('font-size', '11px');
    }

    // Tooltip
    if (showTooltip) {
      const tooltip = g.append('g').attr('class', 'tooltip').style('display', 'none');

      tooltip
        .append('rect')
        .attr('width', 100)
        .attr('height', 40)
        .attr('rx', 4)
        .attr('fill', 'rgba(0,0,0,0.8)');

      const tooltipText = tooltip.append('text').attr('fill', 'white').attr('font-size', '12px');

      g.selectAll('.bar')
        .on('mouseenter', function (event, d: unknown) {
          const data = d as typeof data[0];
          const [x, y] = d3.pointer(event, g.node());

          tooltip
            .attr('transform', `translate(${Math.min(x, innerWidth - 110)},${Math.max(y - 50, 10)})`)
            .style('display', null);

          tooltipText.selectAll('tspan').remove();
          tooltipText.append('tspan').attr('x', 10).attr('y', 18).text(data.label);
          tooltipText.append('tspan').attr('x', 10).attr('y', 32).text(`Value: ${data.value.toLocaleString()}`);
        })
        .on('mouseleave', () => {
          tooltip.style('display', 'none');
        });
    }

    return () => {
      // Cleanup
    };
  };

  const { containerRef } = useD3(renderChart, [data, colors, showValues, showGrid, showTooltip, horizontal, animate], {
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
