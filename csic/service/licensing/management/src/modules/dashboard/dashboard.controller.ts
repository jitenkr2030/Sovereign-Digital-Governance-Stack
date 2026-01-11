import { Controller, Get, UseGuards } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth } from '@nestjs/swagger';
import { DashboardService } from './dashboard.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';

@ApiTags('dashboard')
@Controller('dashboard')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class DashboardController {
  constructor(private readonly dashboardService: DashboardService) {}

  @Get('overview')
  @ApiOperation({ summary: 'Get dashboard overview statistics' })
  @ApiResponse({ status: 200, description: 'Returns comprehensive dashboard data' })
  async getOverview() {
    return this.dashboardService.getOverviewStats();
  }

  @Get('queue')
  @UseGuards(RolesGuard)
  @Roles('admin', 'analyst', 'reviewer')
  @ApiOperation({ summary: 'Get application queue metrics' })
  @ApiResponse({ status: 200, description: 'Returns queue metrics' })
  async getQueueMetrics() {
    return this.dashboardService.getQueueMetrics();
  }

  @Get('compliance')
  @UseGuards(RolesGuard)
  @Roles('admin', 'analyst')
  @ApiOperation({ summary: 'Get compliance metrics' })
  @ApiResponse({ status: 200, description: 'Returns compliance metrics' })
  async getComplianceMetrics() {
    return this.dashboardService.getComplianceMetrics();
  }
}
