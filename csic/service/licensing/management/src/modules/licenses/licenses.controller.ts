import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Query,
  UseGuards,
  Request,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiQuery } from '@nestjs/swagger';
import { LicensesService } from './licenses.service';
import { IssueLicenseDto, SuspendLicenseDto, RevokeLicenseDto } from './dto/license.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { LicenseStatus } from './entities/license-status.enum';

@ApiTags('licenses')
@Controller('licenses')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class LicensesController {
  constructor(private readonly licensesService: LicensesService) {}

  @Get()
  @ApiOperation({ summary: 'List all licenses with filtering' })
  @ApiResponse({ status: 200, description: 'Returns paginated list of licenses' })
  async findAll(
    @Query('status') status?: LicenseStatus,
    @Query('entityId') entityId?: string,
    @Query('licenseTypeId') licenseTypeId?: string,
    @Query('expiringWithinDays') expiringWithinDays?: number,
    @Query('page') page?: number,
    @Query('limit') limit?: number,
  ) {
    return this.licensesService.findAll({
      status,
      entityId,
      licenseTypeId,
      expiringWithinDays,
      page: page ? Number(page) : 1,
      limit: limit ? Number(limit) : 20,
    });
  }

  @Get('expiring')
  @ApiOperation({ summary: 'Get licenses expiring within specified days' })
  @ApiResponse({ status: 200, description: 'Returns expiring licenses' })
  @ApiQuery({ name: 'days', required: false, default: 30 })
  async getExpiring(@Query('days') days: number = 30) {
    return this.licensesService.getExpiringLicenses(days);
  }

  @Get('statistics')
  @UseGuards(RolesGuard)
  @Roles('admin', 'analyst')
  @ApiOperation({ summary: 'Get license statistics' })
  @ApiResponse({ status: 200, description: 'Returns license statistics' })
  async getStatistics() {
    return this.licensesService.getStatistics();
  }

  @Get('number/:licenseNumber')
  @ApiOperation({ summary: 'Get license by license number' })
  @ApiResponse({ status: 200, description: 'Returns the license' })
  @ApiResponse({ status: 404, description: 'License not found' })
  async findByLicenseNumber(@Param('licenseNumber') licenseNumber: string) {
    return this.licensesService.findByLicenseNumber(licenseNumber);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get license by ID' })
  @ApiResponse({ status: 200, description: 'Returns the license' })
  @ApiResponse({ status: 404, description: 'License not found' })
  async findOne(@Param('id') id: string) {
    return this.licensesService.findOne(id);
  }

  @Post('issue')
  @UseGuards(RolesGuard)
  @Roles('admin', 'licensing_officer')
  @ApiOperation({ summary: 'Issue a new license from an approved application' })
  @ApiResponse({ status: 201, description: 'License issued successfully' })
  async issueLicense(@Body() issueDto: IssueLicenseDto, @Request() req) {
    return this.licensesService.issueLicense(
      issueDto.applicationId,
      req.user.id,
      req.user.name,
    );
  }

  @Post(':id/suspend')
  @UseGuards(RolesGuard)
  @Roles('admin', 'licensing_officer')
  @ApiOperation({ summary: 'Suspend a license' })
  @ApiResponse({ status: 200, description: 'License suspended' })
  async suspendLicense(
    @Param('id') id: string,
    @Body() suspendDto: SuspendLicenseDto,
    @Request() req,
  ) {
    return this.licensesService.suspendLicense(
      id,
      suspendDto.reason,
      req.user.name,
    );
  }

  @Post(':id/revoke')
  @UseGuards(RolesGuard)
  @Roles('admin')
  @ApiOperation({ summary: 'Revoke a license' })
  @ApiResponse({ status: 200, description: 'License revoked' })
  async revokeLicense(
    @Param('id') id: string,
    @Body() revokeDto: RevokeLicenseDto,
    @Request() req,
  ) {
    return this.licensesService.revokeLicense(
      id,
      revokeDto.reason,
      req.user.name,
    );
  }

  @Post(':id/renew')
  @UseGuards(RolesGuard)
  @Roles('admin', 'licensing_officer')
  @ApiOperation({ summary: 'Renew a license' })
  @ApiResponse({ status: 200, description: 'License renewed' })
  async renewLicense(@Param('id') id: string) {
    return this.licensesService.renewLicense(id);
  }
}
