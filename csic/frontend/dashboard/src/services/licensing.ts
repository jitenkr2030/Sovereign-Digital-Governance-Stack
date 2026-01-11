import api from './api';
import { License, LicenseApplication, PaginatedResponse, FilterOptions, SortOptions } from '../types';

interface LicenseListParams {
  filters?: FilterOptions;
  sort?: SortOptions;
  page?: number;
  limit?: number;
}

interface LicenseCreateData {
  entityId: string;
  entityName: string;
  entityType: 'VASP' | 'CASP' | 'OTHER';
  licenseType: string;
  jurisdiction: string;
  regulatoryBody: string;
  issueDate: string;
  expiryDate: string;
  conditions?: string[];
  restrictions?: string[];
}

interface LicenseUpdateData {
  status?: string;
  expiryDate?: string;
  conditions?: string[];
  restrictions?: string[];
}

export const licensingService = {
  async getLicenses(params: LicenseListParams = {}): Promise<PaginatedResponse<License>> {
    const { filters, sort, page = 1, limit = 20 } = params;
    
    const queryParams = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });

    if (filters) {
      if (filters.search) queryParams.append('search', filters.search);
      if (filters.status?.length) queryParams.append('status', filters.status.join(','));
      if (filters.type?.length) queryParams.append('type', filters.type.join(','));
      if (filters.jurisdiction?.length) queryParams.append('jurisdiction', filters.jurisdiction.join(','));
      if (filters.dateRange) {
        queryParams.append('startDate', filters.dateRange.start);
        queryParams.append('endDate', filters.dateRange.end);
      }
    }

    if (sort) {
      queryParams.append('sortBy', sort.field);
      queryParams.append('sortOrder', sort.direction);
    }

    return api.get<PaginatedResponse<License>>(`/licenses?${queryParams}`);
  },

  async getLicense(id: string): Promise<License> {
    return api.get<License>(`/licenses/${id}`);
  },

  async createLicense(data: LicenseCreateData): Promise<License> {
    return api.post<License>('/licenses', data);
  },

  async updateLicense(id: string, data: LicenseUpdateData): Promise<License> {
    return api.patch<License>(`/licenses/${id}`, data);
  },

  async deleteLicense(id: string): Promise<void> {
    return api.delete(`/licenses/${id}`);
  },

  async approveLicense(id: string, notes?: string): Promise<License> {
    return api.post<License>(`/licenses/${id}/approve`, { notes });
  },

  async rejectLicense(id: string, reason: string): Promise<License> {
    return api.post<License>(`/licenses/${id}/reject`, { reason });
  },

  async suspendLicense(id: string, reason: string, duration?: string): Promise<License> {
    return api.post<License>(`/licenses/${id}/suspend`, { reason, duration });
  },

  async revokeLicense(id: string, reason: string): Promise<License> {
    return api.post<License>(`/licenses/${id}/revoke`, { reason });
  },

  async renewLicense(id: string, newExpiryDate: string): Promise<License> {
    return api.post<License>(`/licenses/${id}/renew`, { newExpiryDate });
  },

  async getLicenseHistory(id: string): Promise<License['history']> {
    return api.get<License['history']>(`/licenses/${id}/history`);
  },

  async getLicenseDocuments(id: string): Promise<License['documents']> {
    return api.get<License['documents']>(`/licenses/${id}/documents`);
  },

  async uploadLicenseDocument(licenseId: string, file: File, type: string): Promise<License['documents'][0]> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', type);

    return api.post<License['documents'][0]>(`/licenses/${licenseId}/documents`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  async deleteLicenseDocument(licenseId: string, documentId: string): Promise<void> {
    return api.delete(`/licenses/${licenseId}/documents/${documentId}`);
  },

  // Applications
  async getApplications(params?: FilterOptions): Promise<PaginatedResponse<LicenseApplication>> {
    const queryParams = new URLSearchParams();
    if (params?.status?.length) queryParams.append('status', params.status.join(','));
    
    return api.get<PaginatedResponse<LicenseApplication>>(`/applications?${queryParams}`);
  },

  async getApplication(id: string): Promise<LicenseApplication> {
    return api.get<LicenseApplication>(`/applications/${id}`);
  },

  async approveApplication(id: string, data?: { licenseNumber?: string; notes?: string }): Promise<License> {
    return api.post<License>(`/applications/${id}/approve`, data);
  },

  async rejectApplication(id: string, reason: string): Promise<LicenseApplication> {
    return api.post<LicenseApplication>(`/applications/${id}/reject`, { reason });
  },

  // Statistics
  async getLicenseStats(): Promise<{
    total: number;
    active: number;
    pending: number;
    suspended: number;
    revoked: number;
    expiringSoon: number;
  }> {
    return api.get('/licenses/stats');
  },

  async getComplianceStats(): Promise<{
    averageScore: number;
    byStatus: Record<string, number>;
    trend: Array<{ date: string; score: number }>;
  }> {
    return api.get('/licenses/compliance/stats');
  },

  async getJurisdictionStats(): Promise<Array<{
    jurisdiction: string;
    count: number;
    activeCount: number;
  }>> {
    return api.get('/licenses/jurisdictions/stats');
  },
};

export default licensingService;
