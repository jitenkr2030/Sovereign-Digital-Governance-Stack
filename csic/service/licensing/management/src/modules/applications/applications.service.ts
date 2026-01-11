import {
  Injectable,
  NotFoundException,
  BadRequestException,
  ForbiddenException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, Like, In } from 'typeorm';
import { v4 as uuidv4 } from 'uuid';
import { Application, ApplicationStatus, ReviewNote } from './entities/application.entity';
import { CreateApplicationDto, UpdateApplicationDto, ReviewApplicationDto, ApplicationFilterDto } from './dto/create-application.dto';
import { WorkflowService } from '../workflow/workflow.service';
import { KafkaService } from '../kafka/kafka.service';
import { LicenseType } from './entities/license-type.entity';

@Injectable()
export class ApplicationsService {
  constructor(
    @InjectRepository(Application)
    private applicationRepository: Repository<Application>,
    @InjectRepository(LicenseType)
    private licenseTypeRepository: Repository<LicenseType>,
    private workflowService: WorkflowService,
    private kafkaService: KafkaService,
  ) {}

  async create(createDto: CreateApplicationDto, userId: string, userName: string): Promise<Application> {
    // Validate license type exists
    const licenseType = await this.licenseTypeRepository.findOne({
      where: { id: createDto.licenseTypeId, isActive: true },
    });

    if (!licenseType) {
      throw new BadRequestException('Invalid or inactive license type');
    }

    // Create application in DRAFT status
    const application = this.applicationRepository.create({
      ...createDto,
      status: ApplicationStatus.DRAFT,
      licenseType,
      requiredDocuments: licenseType.requirements,
    });

    const saved = await this.applicationRepository.save(application);

    return saved;
  }

  async findAll(filter: ApplicationFilterDto): Promise<{ data: Application[]; total: number; page: number; limit: number }> {
    const { page = 1, limit = 20, status, entityId, licenseTypeId, priority, handlerId } = filter;

    const queryBuilder = this.applicationRepository
      .createQueryBuilder('application')
      .leftJoinAndSelect('application.licenseType', 'licenseType');

    if (status) {
      queryBuilder.andWhere('application.status = :status', { status });
    }

    if (entityId) {
      queryBuilder.andWhere('application.entityId = :entityId', { entityId });
    }

    if (licenseTypeId) {
      queryBuilder.andWhere('application.licenseTypeId = :licenseTypeId', { licenseTypeId });
    }

    if (priority) {
      queryBuilder.andWhere('application.priority = :priority', { priority });
    }

    if (handlerId) {
      queryBuilder.andWhere('application.currentHandlerId = :handlerId', { handlerId });
    }

    queryBuilder
      .orderBy('application.priority', 'DESC')
      .addOrderBy('application.createdAt', 'ASC')
      .skip((page - 1) * limit)
      .take(limit);

    const [data, total] = await queryBuilder.getManyAndCount();

    return { data, total, page, limit };
  }

  async findOne(id: string): Promise<Application> {
    const application = await this.applicationRepository.findOne({
      where: { id },
      relations: ['licenseType', 'documents'],
    });

    if (!application) {
      throw new NotFoundException(`Application with ID ${id} not found`);
    }

    return application;
  }

  async update(id: string, updateDto: UpdateApplicationDto, userId: string): Promise<Application> {
    const application = await this.findOne(id);

    // Can only update applications in DRAFT status
    if (application.status !== ApplicationStatus.DRAFT) {
      throw new BadRequestException('Can only update applications in DRAFT status');
    }

    Object.assign(application, updateDto);
    await this.applicationRepository.save(application);

    return this.findOne(id);
  }

  async submit(id: string, userId: string, userName: string): Promise<Application> {
    const application = await this.findOne(id);

    // Validate transition
    const canTransition = this.workflowService.canTransition(
      application.status,
      ApplicationStatus.SUBMITTED,
    );

    if (!canTransition) {
      throw new BadRequestException(
        `Cannot transition from ${application.status} to SUBMITTED`,
      );
    }

    // Validate required documents
    const missingDocuments = this.validateRequiredDocuments(application);
    if (missingDocuments.length > 0) {
      throw new BadRequestException({
        message: 'Missing required documents',
        missingDocuments,
      });
    }

    // Update status
    application.status = ApplicationStatus.SUBMITTED;
    application.submittedAt = new Date();
    await this.applicationRepository.save(application);

    // Publish Kafka event
    await this.kafkaService.publishLicenseEvent('license.application.submitted', {
      applicationId: application.id,
      entityId: application.entityId,
      entityName: application.entityName,
      licenseType: application.licenseType?.name,
      submittedAt: application.submittedAt,
    });

    return this.findOne(id);
  }

  async review(id: string, reviewDto: ReviewApplicationDto, handlerId: string, handlerName: string): Promise<Application> {
    const application = await this.findOne(id);

    // Validate transition
    const canTransition = this.workflowService.canTransition(
      application.status,
      reviewDto.status,
    );

    if (!canTransition) {
      throw new BadRequestException(
        `Cannot transition from ${application.status} to ${reviewDto.status}`,
      );
    }

    // Handle status-specific logic
    if (reviewDto.status === ApplicationStatus.UNDER_REVIEW && !application.reviewStartedAt) {
      application.reviewStartedAt = new Date();
    }

    if (reviewDto.status === ApplicationStatus.APPROVED || reviewDto.status === ApplicationStatus.REJECTED) {
      application.decision = reviewDto.status;
      application.decisionReason = reviewDto.decisionReason;
      application.decisionAt = new Date();
    }

    // Add review note
    if (reviewDto.notes) {
      const note: ReviewNote = {
        id: uuidv4(),
        handlerId,
        handlerName,
        content: reviewDto.notes,
        createdAt: new Date(),
        isInternal: reviewDto.isInternal || false,
      };

      application.reviewNotes = [...(application.reviewNotes || []), note];
    }

    // Update handler
    application.currentHandlerId = handlerId;
    application.currentHandlerName = handlerName;

    application.status = reviewDto.status;
    await this.applicationRepository.save(application);

    // Publish Kafka event
    const eventType = reviewDto.status === ApplicationStatus.APPROVED
      ? 'license.application.approved'
      : reviewDto.status === ApplicationStatus.REJECTED
        ? 'license.application.rejected'
        : `license.application.review.${reviewDto.status.toLowerCase()}`;

    await this.kafkaService.publishLicenseEvent(eventType, {
      applicationId: application.id,
      entityId: application.entityId,
      entityName: application.entityName,
      newStatus: reviewDto.status,
      handlerId,
      handlerName,
      decisionReason: reviewDto.decisionReason,
    });

    return this.findOne(id);
  }

  async getWorkflowStatus(id: string): Promise<any> {
    const application = await this.findOne(id);

    const currentState = this.workflowService.getState(application.status);
    const allowedTransitions = this.workflowService.getAllowedTransitions(application.status);

    return {
      applicationId: id,
      currentStatus: application.status,
      state: currentState,
      allowedTransitions,
      workflowHistory: application.reviewNotes,
      nextSteps: allowedTransitions.map(t => this.getStepDescription(t)),
    };
  }

  async getQueueStats(): Promise<any> {
    const stats = await this.applicationRepository
      .createQueryBuilder('application')
      .select('application.status', 'status')
      .addSelect('COUNT(*)', 'count')
      .groupBy('application.status')
      .getRawMany();

    const priorityCounts = await this.applicationRepository
      .createQueryBuilder('application')
      .select('application.priority', 'priority')
      .addSelect('COUNT(*)', 'count')
      .where('application.status IN (:...statuses)', {
        statuses: [ApplicationStatus.SUBMITTED, ApplicationStatus.UNDER_REVIEW],
      })
      .groupBy('application.priority')
      .getRawMany();

    return {
      byStatus: stats.reduce((acc, s) => ({ ...acc, [s.status]: parseInt(s.count) }), {}),
      byPriority: priorityCounts.reduce((acc, p) => ({ ...acc, [p.priority]: parseInt(p.count) }), {}),
      totalPending: stats
        .filter(s => [ApplicationStatus.SUBMITTED, ApplicationStatus.UNDER_REVIEW].includes(s.status as ApplicationStatus))
        .reduce((sum, s) => sum + parseInt(s.count), 0),
    };
  }

  private validateRequiredDocuments(application: Application): string[] {
    const uploadedDocTypes = (application.documents || [])
      .filter(d => d.status !== 'REJECTED')
      .map(d => d.documentType);

    return (application.requiredDocuments || [])
      .filter(required => !uploadedDocTypes.includes(required));
  }

  private getStepDescription(status: ApplicationStatus): string {
    const descriptions: Record<ApplicationStatus, string> = {
      [ApplicationStatus.DRAFT]: 'Complete application details and upload required documents',
      [ApplicationStatus.SUBMITTED]: 'Application submitted for regulatory review',
      [ApplicationStatus.UNDER_REVIEW]: 'Regulator is reviewing the application',
      [ApplicationStatus.PENDING_INFO]: 'Additional information required from applicant',
      [ApplicationStatus.APPROVED]: 'License approved',
      [ApplicationStatus.REJECTED]: 'Application rejected',
    };

    return descriptions[status] || status;
  }
}
