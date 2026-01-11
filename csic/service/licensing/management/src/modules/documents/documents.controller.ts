import {
  Controller,
  Get,
  Post,
  Delete,
  Body,
  Param,
  Query,
  UseGuards,
  UseInterceptors,
  UploadedFile,
  Request,
} from '@nestjs/common';
import { FileInterceptor } from '@nestjs/platform-express';
import { ApiTags, ApiOperation, ApiResponse, ApiBearerAuth, ApiConsumes, ApiBody } from '@nestjs/swagger';
import { DocumentsService } from './documents.service';
import { VerifyDocumentDto, RejectDocumentDto, UploadDocumentDto } from './dto/document.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { DocumentType } from './entities/document.entity';

@ApiTags('documents')
@Controller('documents')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class DocumentsController {
  constructor(private readonly documentsService: DocumentsService) {}

  @Post('upload')
  @UseGuards(RolesGuard)
  @Roles('admin', 'analyst', 'reviewer', 'applicant')
  @ApiOperation({ summary: 'Upload a document for an application' })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    schema: {
      type: 'object',
      properties: {
        file: { type: 'string', format: 'binary' },
        applicationId: { type: 'string' },
        documentType: { type: 'string', enum: Object.values(DocumentType) },
      },
    },
  })
  @UseInterceptors(FileInterceptor('file'))
  async uploadDocument(
    @UploadedFile() file: Express.Multer.File,
    @Body() uploadDto: UploadDocumentDto,
    @Request() req,
  ) {
    return this.documentsService.uploadDocument(
      uploadDto.applicationId,
      file,
      uploadDto.documentType as DocumentType,
      req.user.id,
      req.user.name,
    );
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get document metadata' })
  @ApiResponse({ status: 200, description: 'Returns document metadata' })
  @ApiResponse({ status: 404, description: 'Document not found' })
  async getDocument(@Param('id') id: string) {
    return this.documentsService.getDocument(id);
  }

  @Get(':id/download')
  @ApiOperation({ summary: 'Get signed download URL for document' })
  @ApiResponse({ status: 200, description: 'Returns signed URL' })
  async getDownloadUrl(@Param('id') id: string) {
    const url = await this.documentsService.getSignedDownloadUrl(id);
    return { url, expiresIn: '1 hour' };
  }

  @Post(':id/verify')
  @UseGuards(RolesGuard)
  @Roles('admin', 'reviewer')
  @ApiOperation({ summary: 'Verify a document' })
  @ApiResponse({ status: 200, description: 'Document verified' })
  async verifyDocument(
    @Param('id') id: string,
    @Body() verifyDto: VerifyDocumentDto,
    @Request() req,
  ) {
    return this.documentsService.verifyDocument(
      id,
      req.user.id,
      req.user.name,
      verifyDto.notes,
    );
  }

  @Post(':id/reject')
  @UseGuards(RolesGuard)
  @Roles('admin', 'reviewer')
  @ApiOperation({ summary: 'Reject a document' })
  @ApiResponse({ status: 200, description: 'Document rejected' })
  async rejectDocument(
    @Param('id') id: string,
    @Body() rejectDto: RejectDocumentDto,
    @Request() req,
  ) {
    return this.documentsService.rejectDocument(id, rejectDto.reason, req.user.name);
  }

  @Get(':id/integrity')
  @ApiOperation({ summary: 'Verify document integrity via checksum' })
  @ApiResponse({ status: 200, description: 'Returns integrity check result' })
  async verifyIntegrity(@Param('id') id: string) {
    return this.documentsService.verifyIntegrity(id);
  }

  @Get('application/:applicationId')
  @ApiOperation({ summary: 'Get all documents for an application' })
  @ApiResponse({ status: 200, description: 'Returns document list' })
  async getDocumentsByApplication(@Param('applicationId') applicationId: string) {
    return this.documentsService.getDocumentsByApplication(applicationId);
  }

  @Get('pending/verification')
  @UseGuards(RolesGuard)
  @Roles('admin', 'reviewer')
  @ApiOperation({ summary: 'Get documents pending verification' })
  @ApiResponse({ status: 200, description: 'Returns pending documents' })
  async getPendingVerification() {
    return this.documentsService.getPendingVerification();
  }

  @Delete(':id')
  @UseGuards(RolesGuard)
  @Roles('admin')
  @ApiOperation({ summary: 'Delete a document' })
  @ApiResponse({ status: 200, description: 'Document deleted' })
  async deleteDocument(@Param('id') id: string) {
    await this.documentsService.deleteDocument(id);
    return { message: 'Document deleted successfully' };
  }
}
