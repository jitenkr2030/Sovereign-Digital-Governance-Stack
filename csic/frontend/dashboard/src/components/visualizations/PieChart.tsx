import { useD3 } from '../../hooks/useD3';
import * as d3 from 'd3';

interface PieChartProps {
  data: Array<{
    name: string;
    value: number;
    color?: string;
  }>;
  width?: number;
  height?: number;
  colors?: string[];
  innerRadius?: number;
  showLabels?: boolean;
  showTooltip?: boolean;
  showLegend?: boolean;
  animate?: boolean;
  donut?: boolean;
}

export default function PieChart({
  data,
  width = 400,
  height = 400,
  colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'],
  innerRadius = 0,
  showLabels = true,
  showTooltip = true,
  showLegend = true,
  animate = true,
  donut = false,
}: PieChartProps) {
  const renderChart = (svg: d3.Selection<SVGSVGElement, unknown, null, undefined>, dimensions: { width: number; height: number }) => {
    const { width: w, height: h } = dimensions;
    const radius = Math.min(w, h) / 2 - 20;
    const centerX = w / 2;
    const centerY = h / 2;

    const g = svg.append('g').attr('transform', `translate(${centerX},${centerY})`);

    const colorScale = d3.scaleOrdinal<string>().domain(data.map((d) => d.name)).range(colors);

    // Calculate total
    const total = d3.sum(data, (d) => d.value);

    // Create pie generator
    const pie = d3
      .pie<typeof data[0]>()
      .value((d) => d.value)
      .sort(null)
      .padAngle(0.02);

    // Create arc generator
    const arc = d3
      .arc<d3.PieArcDatum<typeof data[0]>>()
      .innerRadius(donut ? radius * 0.5 : innerRadius)
      .outerRadius(radius)
      .cornerRadius(4);

    // Create hover arc
    const hoverArc = d3
      .arc<d3.PieArcDatum<typeof data[0]>>()
      .innerRadius(donut ? radius * 0.5 : innerRadius)
      .outerRadius(radius + 10)
      .cornerRadius(4);

    // Create label arc
    const labelArc = d3
      .arc<d3.PieArcDatum<typeof data[0]>>()
      .innerRadius(radius * 0.7)
      .outerRadius(radius * 0.7);

    // Create pie data
    const pieData = pie(data);

    // Create tooltip
    const tooltip = svg
      .append('g')
      .attr('class', 'tooltip')
      .style('display', 'none');

    tooltip
      .append('rect')
      .attr('width', 120)
      .attr('height', 50)
      .attr('rx', 4)
      .attr('fill', 'rgba(0,0,0,0.8)');

    const tooltipText = tooltip.append('text').attr('fill', 'white').attr('font-size', '12px');

    // Draw arcs
    const arcs = g
      .selectAll('.arc')
      .data(pieData)
      .join('g')
      .attr('class', 'arc');

    const paths = arcs
      .append('path')
      .attr('fill', (d) => d.data.color || colorScale(d.data.name))
      .attr('stroke', 'white')
      .attr('stroke-width', 2)
      .style('cursor', 'pointer');

    if (animate) {
      paths
        .transition()
        .duration(1000)
        .attrTween('d', function (d) {
          const interpolate = d3.interpolate({ startAngle: 0, endAngle: 0 }, d);
          return (t) => arc(interpolate(t)) || '';
        });
    } else {
      paths.attr('d', arc);
    }

    // Hover effects
    paths
      .on('mouseenter', function (event, d) {
        d3.select(this)
          .transition()
          .duration(200)
          .attr('d', hoverArc(d) || '');

        const [x, y] = d3.pointer(event, svg.node());

        tooltip
          .attr('transform', `translate(${Math.min(x, w - 130)},${Math.max(y - 60, 10)})`)
          .style('display', null);

        tooltipText.selectAll('tspan').remove();
        tooltipText.append('tspan').attr('x', 10).attr('y', 20).text(d.data.name);
        tooltipText
          .append('tspan')
          .attr('x', 10)
          .attr('y', 36)
          .text(`${((d.data.value / total) * 100).toFixed(1)}% (${d.data.value.toLocaleString()})`);
      })
      .on('mouseleave', function (_, d) {
        d3.select(this)
          .transition()
          .duration(200)
          .attr('d', arc(d) || '');

        tooltip.style('display', 'none');
      });

    // Labels
    if (showLabels) {
      arcs
        .append('text')
        .attr('transform', (d) => `translate(${labelArc.centroid(d)})`)
        .attr('text-anchor', 'middle')
        .attr('fill', 'white')
        .attr('font-size', '11px')
        .attr('font-weight', '500')
        .style('opacity', animate ? 0 : 1)
        .text((d) => (d.endAngle - d.startAngle > 0.3 ? `${((d.data.value / total) * 100).toFixed(0)}%` : ''))
        .transition()
        .delay(animate ? 1000 : 0)
        .duration(300)
        .style('opacity', 1);
    }

    // Center text for donut chart
    if (donut) {
      const centerGroup = g.append('g').attr('class', 'center-text');

      centerGroup
        .append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '-0.2em')
        .attr('fill', 'currentColor')
        .attr('font-size', '24px')
        .attr('font-weight', 'bold')
        .text(total.toLocaleString());

      centerGroup
        .append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '1.2em')
        .attr('fill', 'currentColor')
        .attr('font-size', '12px')
        .attr('opacity', 0.7)
        .text('Total');
    }

    // Legend
    if (showLegend && data.length <= 8) {
      const legend = svg
        .append('g')
        .attr('transform', `translate(${w - 100}, 20)`);

      data.forEach((item, i) => {
        const legendItem = legend
          .append('g')
          .attr('transform', `translate(0, ${i * 20})`);

        legendItem
          .append('rect')
          .attr('width', 12)
          .attr('height', 12)
          .attr('rx', 2)
          .attr('fill', item.color || colorScale(item.name));

        legendItem
          .append('text')
          .attr('x', 18)
          .attr('y', 10)
          .attr('fill', 'currentColor')
          .attr('font-size', '11px')
          .text(item.name.length > 10 ? item.name.substring(0, 10) + '...' : item.name);
      });
    }

    return () => {
      // Cleanup
    };
  };

  const { containerRef } = useD3(renderChart, [data, colors, innerRadius, showLabels, showTooltip, showLegend, animate, donut], {
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
