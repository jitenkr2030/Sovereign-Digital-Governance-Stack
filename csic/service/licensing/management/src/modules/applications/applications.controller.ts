import {
  Controller,
  Get,
  Post,
  Patch,
  Body,
  Param,
  Query,
  UseGuards,
  Request,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiQuery } from '@nestjs/swagger';
import { ApplicationsService } from './applications.service';
import { CreateApplicationDto, UpdateApplicationDto, ReviewApplicationDto, SubmitApplicationDto, ApplicationFilterDto } from './dto/create-application.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';

@ApiTags('applications')
@Controller('applications')
export class ApplicationsController {
  constructor(private readonly applicationsService: ApplicationsService) {}

  @Post()
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Create a new license application' })
  @ApiResponse({ status: 201, description: 'Application created successfully' })
  @ApiResponse({ status: 400, description: 'Invalid input' })
  async create(
    @Body() createDto: CreateApplicationDto,
    @Request() req,
  ) {
    return this.applicationsService.create(
      createDto,
      req.user.id,
      req.user.name,
    );
  }

  @Get()
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'List all applications with filtering' })
  @ApiResponse({ status: 200, description: 'Returns paginated list of applications' })
  async findAll(@Query() filter: ApplicationFilterDto) {
    return this.applicationsService.findAll(filter);
  }

  @Get('queue/stats')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'analyst', 'reviewer')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Get application queue statistics' })
  @ApiResponse({ status: 200, description: 'Returns queue statistics' })
  async getQueueStats() {
    return this.applicationsService.getQueueStats();
  }

  @Get(':id')
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Get application by ID' })
  @ApiResponse({ status: 200, description: 'Returns the application' })
  @ApiResponse({ status: 404, description: 'Application not found' })
  async findOne(@Param('id') id: string) {
    return this.applicationsService.findOne(id);
  }

  @Get(':id/workflow')
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Get application workflow status and allowed transitions' })
  @ApiResponse({ status: 200, description: 'Returns workflow information' })
  async getWorkflowStatus(@Param('id') id: string) {
    return this.applicationsService.getWorkflowStatus(id);
  }

  @Patch(':id')
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Update application (DRAFT status only)' })
  @ApiResponse({ status: 200, description: 'Application updated' })
  @ApiResponse({ status: 400, description: 'Cannot update application' })
  async update(
    @Param('id') id: string,
    @Body() updateDto: UpdateApplicationDto,
    @Request() req,
  ) {
    return this.applicationsService.update(id, updateDto, req.user.id);
  }

  @Post(':id/submit')
  @UseGuards(JwtAuthGuard)
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Submit application for review' })
  @ApiResponse({ status: 200, description: 'Application submitted' })
  @ApiResponse({ status: 400, description: 'Submission failed' })
  async submit(
    @Param('id') id: string,
    @Body() submitDto: SubmitApplicationDto,
    @Request() req,
  ) {
    return this.applicationsService.submit(id, req.user.id, req.user.name);
  }

  @Post(':id/review')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'reviewer')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Review and update application status' })
  @ApiResponse({ status: 200, description: 'Review recorded' })
  async review(
    @Param('id') id: string,
    @Body() reviewDto: ReviewApplicationDto,
    @Request() req,
  ) {
    return this.applicationsService.review(
      id,
      reviewDto,
      req.user.id,
      req.user.name,
    );
  }
}
