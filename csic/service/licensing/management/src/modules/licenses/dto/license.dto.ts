import { IsString, IsUUID, IsOptional, IsEnum, IsNotEmpty } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { LicenseStatus } from '../entities/license-status.enum';

export class IssueLicenseDto {
  @ApiProperty({ description: 'Application ID to issue license for' })
  @IsUUID()
  applicationId: string;
}

export class SuspendLicenseDto {
  @ApiProperty({ description: 'Reason for suspension' })
  @IsString()
  @IsNotEmpty()
  reason: string;
}

export class RevokeLicenseDto {
  @ApiProperty({ description: 'Reason for revocation' })
  @IsString()
  @IsNotEmpty()
  reason: string;
}

export class LicenseFilterDto {
  @ApiPropertyOptional({ description: 'Filter by status', enum: LicenseStatus })
  @IsEnum(LicenseStatus)
  @IsOptional()
  status?: LicenseStatus;

  @ApiPropertyOptional({ description: 'Filter by entity ID' })
  @IsUUID()
  @IsOptional()
  entityId?: string;

  @ApiPropertyOptional({ description: 'Filter by license type ID' })
  @IsUUID()
  @IsOptional()
  licenseTypeId?: string;

  @ApiPropertyOptional({ description: 'Expiring within days' })
  @IsOptional()
  expiringWithinDays?: number;

  @ApiPropertyOptional({ description: 'Page number', default: 1 })
  @IsOptional()
  page?: number;

  @ApiPropertyOptional({ description: 'Items per page', default: 20 })
  @IsOptional()
  limit?: number;
}
