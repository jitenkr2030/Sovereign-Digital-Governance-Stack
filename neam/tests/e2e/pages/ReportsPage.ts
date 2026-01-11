import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Reports page
 * Handles all report generation and management interactions
 */
export class ReportsPage {
  readonly page: Page;
  
  // Page header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;
  
  // Action buttons
  readonly generateReportButton: Locator;
  readonly scheduleReportButton: Locator;
  readonly exportButton: Locator;
  readonly importButton: Locator;
  
  // Report templates
  readonly templatesSection: Locator;
  readonly templateCards: Locator;
  readonly weeklyIntelligenceTemplate: Locator;
  readonly monthlyThreatTemplate: Locator;
  readonly incidentSummaryTemplate: Locator;
  readonly customReportTemplate: Locator;
  
  // Saved reports
  readonly savedReportsSection: Locator;
  readonly savedReportsList: Locator;
  readonly savedReportItems: Locator;
  
  // Search and filters
  readonly searchInput: Locator;
  readonly dateRangeFilter: Locator;
  readonly typeFilter: Locator;
  readonly statusFilter: Locator;
  readonly sortBy: Locator;
  
  // Report details
  readonly reportDetails: Locator;
  readonly reportTitle: Locator;
  readonly reportContent: Locator;
  readonly reportCharts: Locator;
  readonly reportTables: Locator;
  
  // Report actions
  readonly viewReportButton: Locator;
  readonly downloadReportButton: Locator;
  readonly shareReportButton: Locator;
  readonly scheduleReportButtonItem: Locator;
  readonly deleteReportButton: Locator;
  
  // Status indicators
  readonly generatingStatusBadge: Locator;
  readonly completedStatusBadge: Locator;
  readonly failedStatusBadge: Locator;
  readonly scheduledStatusBadge: Locator;
  
  // Report types
  readonly weeklyReportType: Locator;
  readonly monthlyReportType: Locator;
  readonly quarterlyReportType: Locator;
  readonly annualReportType: Locator;
  readonly customReportType: Locator;
  
  // Date range picker
  readonly startDateInput: Locator;
  readonly endDateInput: Locator;
  readonly applyDateRangeButton: Locator;
  readonly quickDateRanges: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Page header
    this.pageTitle = page.locator('[data-testid="page-title"], h1:has-text("Reports"), .page-title');
    this.pageDescription = page.locator('[data-testid="page-description"], .page-description');
    
    // Actions
    this.generateReportButton = page.locator('[data-testid="generate-report-button"], button:has-text("Generate Report"), button:has-text("New Report"), [class*="generate"]');
    this.scheduleReportButton = page.locator('[data-testid="schedule-report-button"], button:has-text("Schedule Report"), [class*="schedule"]');
    this.exportButton = page.locator('[data-testid="export-button"], button:has-text("Export"), [class*="export"]');
    this.importButton = page.locator('[data-testid="import-button"], button:has-text("Import"), [class*="import"]');
    
    // Templates
    this.templatesSection = page.locator('[data-testid="templates-section"], .templates-section, [class*="templates"]');
    this.templateCards = page.locator('[data-testid="template-card"], .template-card, [class*="template"]');
    this.weeklyIntelligenceTemplate = page.locator('[data-testid="template-weekly-intelligence"], [class*="template"]:has-text("Weekly Intelligence")');
    this.monthlyThreatTemplate = page.locator('[data-testid="template-monthly-threat"], [class*="template"]:has-text("Monthly Threat")');
    this.incidentSummaryTemplate = page.locator('[data-testid="template-incident-summary"], [class*="template"]:has-text("Incident Summary")');
    this.customReportTemplate = page.locator('[data-testid="template-custom"], [class*="template"]:has-text("Custom Report")');
    
    // Saved reports
    this.savedReportsSection = page.locator('[data-testid="saved-reports-section"], .saved-reports, [class*="saved-reports"]');
    this.savedReportsList = page.locator('[data-testid="saved-reports-list"], .reports-list, [class*="reports-list"]');
    this.savedReportItems = page.locator('[data-testid="report-item"], .report-item, [class*="report-item"]');
    
    // Filters
    this.searchInput = page.locator('[data-testid="search-input"], input[type="search"]');
    this.dateRangeFilter = page.locator('[data-testid="date-range-filter"], [class*="date-filter"]');
    this.typeFilter = page.locator('[data-testid="type-filter"], [class*="type-filter"]');
    this.statusFilter = page.locator('[data-testid="status-filter"], [class*="status-filter"]');
    this.sortBy = page.locator('[data-testid="sort-by"], [class*="sort-by"]');
    
    // Report details
    this.reportDetails = page.locator('[data-testid="report-details"], .report-details, [class*="report-details"]');
    this.reportTitle = page.locator('[data-testid="report-title"], h2, [class*="report-title"]');
    this.reportContent = page.locator('[data-testid="report-content"], .report-content, [class*="content"]');
    this.reportCharts = page.locator('[data-testid="report-charts"], .report-charts, [class*="chart"]');
    this.reportTables = page.locator('[data-testid="report-tables"], .report-tables, [class*="table"]');
    
    // Report actions
    this.viewReportButton = page.locator('[data-testid="view-button"], button:has-text("View")');
    this.downloadReportButton = page.locator('[data-testid="download-button"], button:has-text("Download"), [class*="download"]');
    this.shareReportButton = page.locator('[data-testid="share-button"], button:has-text("Share"), [class*="share"]');
    this.scheduleReportButtonItem = page.locator('[data-testid="schedule-button"], button:has-text("Schedule")');
    this.deleteReportButton = page.locator('[data-testid="delete-button"], button:has-text("Delete")');
    
    // Status badges
    this.generatingStatusBadge = page.locator('[data-testid="status-generating"], [class*="status"]:has-text("Generating")');
    this.completedStatusBadge = page.locator('[data-testid="status-completed"], [class*="status"]:has-text("Completed")');
    this.failedStatusBadge = page.locator('[data-testid="status-failed"], [class*="status"]:has-text("Failed")');
    this.scheduledStatusBadge = page.locator('[data-testid="status-scheduled"], [class*="status"]:has-text("Scheduled")');
    
    // Date range
    this.startDateInput = page.locator('[data-testid="start-date"], input[type="date"][name*="start"]');
    this.endDateInput = page.locator('[data-testid="end-date"], input[type="date"][name*="end"]');
    this.applyDateRangeButton = page.locator('[data-testid="apply-date-range"], button:has-text("Apply"), button:has-text("Update")');
    this.quickDateRanges = page.locator('[data-testid="quick-date-range"], [class*="quick-date"]:not([class*="filter"])');
  }

  /**
   * Navigate to reports page
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/reports`, { waitUntil: 'networkidle' });
  }

  /**
   * Wait for page to fully load
   */
  async waitForPageLoad(): Promise<void> {
    await expect(this.templatesSection.or(this.savedReportsSection)).toBeVisible({ timeout: 10000 });
  }

  /**
   * Get page title
   */
  async getPageTitle(): Promise<string> {
    const titleElement = this.pageTitle.first();
    await expect(titleElement).toBeVisible();
    return await titleElement.textContent() || '';
  }

  /**
   * Get saved reports count
   */
  async getSavedReportsCount(): Promise<number> {
    return await this.savedReportItems.count();
  }

  /**
   * Click generate report button
   */
  async clickGenerateReport(): Promise<void> {
    await this.generateReportButton.click();
  }

  /**
   * Select a report template
   * @param templateName - Name of the template to select
   */
  async selectTemplate(templateName: string): Promise<void> {
    const template = this.templateCards.filter({ hasText: new RegExp(templateName, 'i') });
    await expect(template).toBeVisible();
    await template.click();
  }

  /**
   * Select weekly intelligence template
   */
  async selectWeeklyIntelligenceTemplate(): Promise<void> {
    await this.weeklyIntelligenceTemplate.click();
  }

  /**
   * Select monthly threat template
   */
  async selectMonthlyThreatTemplate(): Promise<void> {
    await this.monthlyThreatTemplate.click();
  }

  /**
   * Select custom report template
   */
  async selectCustomReportTemplate(): Promise<void> {
    await this.customReportTemplate.click();
  }

  /**
   * Set date range for report
   * @param startDate - Start date
   * @param endDate - End date
   */
  async setDateRange(startDate: string, endDate: string): Promise<void> {
    await this.startDateInput.fill(startDate);
    await this.endDateInput.fill(endDate);
    await this.applyDateRangeButton.click();
  }

  /**
   * Select quick date range
   * @param rangeName - Name of the quick range (e.g., "Last 7 days", "Last 30 days")
   */
  async selectQuickDateRange(rangeName: string): Promise<void> {
    const rangeOption = this.quickDateRanges.filter({ hasText: new RegExp(rangeName, 'i') });
    await rangeOption.click();
  }

  /**
   * Search for reports
   * @param query - Search query
   */
  async search(query: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(query);
    await this.searchInput.press('Enter');
  }

  /**
   * Filter by report type
   * @param type - Report type
   */
  async filterByType(type: string): Promise<void> {
    await this.typeFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(type, 'i') });
    await option.click();
  }

  /**
   * Filter by status
   * @param status - Report status
   */
  async filterByStatus(status: string): Promise<void> {
    await this.statusFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(status, 'i') });
    await option.click();
  }

  /**
   * View saved report
   * @param reportTitle - Title of the report
   */
  async viewReport(reportTitle: string): Promise<void> {
    const reportItem = this.getReportByTitle(reportTitle);
    const viewButton = reportItem.locator(this.viewReportButton);
    await expect(viewButton).toBeVisible();
    await viewButton.click();
  }

  /**
   * Download report
   * @param reportTitle - Title of the report
   * @param format - Download format (pdf, csv, xlsx)
   */
  async downloadReport(reportTitle: string, format: 'pdf' | 'csv' | 'xlsx' = 'pdf'): Promise<void> {
    const reportItem = this.getReportByTitle(reportTitle);
    const downloadButton = reportItem.locator(this.downloadReportButton);
    await expect(downloadButton).toBeVisible();
    await downloadButton.click();
    
    // Select format from dropdown
    const formatOption = this.page.locator('[role="option"], [role="menuitem"]', { hasText: new RegExp(format, 'i') });
    await formatOption.click();
  }

  /**
   * Share report
   * @param reportTitle - Title of the report
   * @param recipientEmail - Email to share with
   */
  async shareReport(reportTitle: string, recipientEmail: string): Promise<void> {
    const reportItem = this.getReportByTitle(reportTitle);
    const shareButton = reportItem.locator(this.shareReportButton);
    await shareButton.click();
    
    // Fill in share dialog
    const emailInput = this.page.locator('input[type="email"], input[name="email"]');
    await emailInput.fill(recipientEmail);
    
    const sendButton = this.page.locator('button:has-text("Share"), button:has-text("Send")');
    await sendButton.click();
  }

  /**
   * Delete report
   * @param reportTitle - Title of the report
   */
  async deleteReport(reportTitle: string): Promise<void> {
    const reportItem = this.getReportByTitle(reportTitle);
    const deleteButton = reportItem.locator(this.deleteReportButton);
    await deleteButton.click();
    
    // Confirm deletion
    const confirmButton = this.page.locator('button:has-text("Delete"), button:has-text("Confirm"), [data-testid="confirm-delete"]');
    await confirmButton.click();
  }

  /**
   * Schedule a report
   * @param reportTitle - Title of the report
   * @param schedule - Schedule configuration
   */
  async scheduleReport(reportTitle: string, schedule: { frequency: string; time: string; recipients: string[] }): Promise<void> {
    const reportItem = this.getReportByTitle(reportTitle);
    const scheduleButton = reportItem.locator(this.scheduleReportButtonItem);
    await scheduleButton.click();
    
    // Configure schedule
    const frequencyOption = this.page.locator('[role="listbox"]', { hasText: new RegExp(schedule.frequency, 'i') });
    await frequencyOption.click();
    
    const timeInput = this.page.locator('input[type="time"]');
    await timeInput.fill(schedule.time);
    
    // Add recipients
    for (const recipient of schedule.recipients) {
      const recipientInput = this.page.locator('input[placeholder*="email"], input[name*="recipient"]');
      await recipientInput.fill(recipient);
      await recipientInput.press('Enter');
    }
    
    const saveButton = this.page.locator('button:has-text("Save Schedule"), button:has-text("Save")');
    await saveButton.click();
  }

  /**
   * Get report by title
   * @param reportTitle - Title to search for
   */
  private getReportByTitle(reportTitle: string): Locator {
    return this.savedReportItems.filter({ has: this.page.locator('text', { hasText: new RegExp(reportTitle, 'i') }) });
  }

  /**
   * Get report status
   * @param reportTitle - Title of the report
   */
  async getReportStatus(reportTitle: string): Promise<string> {
    const reportItem = this.getReportByTitle(reportTitle);
    const statusElement = reportItem.locator('[class*="status"], [class*="badge"]').first();
    await expect(statusElement).toBeVisible();
    return await statusElement.textContent() || '';
  }

  /**
   * Check if report exists
   * @param reportTitle - Title to search for
   */
  async reportExists(reportTitle: string): Promise<boolean> {
    const reportItem = this.getReportByTitle(reportTitle);
    return await reportItem.isVisible();
  }

  /**
   * Wait for report generation to complete
   * @param reportTitle - Title of the report
   * @param timeout - Maximum wait time in ms
   */
  async waitForGenerationComplete(reportTitle: string, timeout: number = 120000): Promise<void> {
    await this.page.waitForFunction(
      async (title: string) => {
        const items = document.querySelectorAll('[class*="report-item"]');
        for (const item of items) {
          if (item.textContent?.includes(title)) {
            const status = item.querySelector('[class*="status"]');
            if (status && !status.textContent?.includes('Generating') && !status.textContent?.includes('In Progress')) {
              return true;
            }
          }
        }
        return false;
      },
      reportTitle,
      { timeout }
    );
  }

  /**
   * Get report content
   */
  async getReportContent(): Promise<string> {
    await expect(this.reportContent).toBeVisible();
    return await this.reportContent.textContent() || '';
  }

  /**
   * Export all reports
   * @param format - Export format
   */
  async exportAllReports(format: 'csv' | 'xlsx' = 'csv'): Promise<void> {
    await this.exportButton.click();
    const formatOption = this.page.locator('[role="option"]', { hasText: new RegExp(format, 'i') });
    await formatOption.click();
  }

  /**
   * Sort reports by
   * @param sortOption - Sort option (date, title, status)
   */
  async sortByOption(sortOption: string): Promise<void> {
    await this.sortBy.click();
    const option = this.page.locator('option', { hasText: new RegExp(sortOption, 'i') });
    await option.click();
  }

  /**
   * Create custom report with all options
   * @param options - Custom report options
   */
  async createCustomReport(options: {
    title: string;
    type: string;
    startDate: string;
    endDate: string;
    sections: string[];
    includeCharts: boolean;
    includeTables: boolean;
  }): Promise<void> {
    await this.selectCustomReportTemplate();
    
    // Fill in custom report form
    const titleInput = this.page.locator('input[name="title"], input[placeholder*="Title"]');
    await titleInput.fill(options.title);
    
    await this.setDateRange(options.startDate, options.endDate);
    
    // Select sections
    for (const section of options.sections) {
      const sectionCheckbox = this.page.locator('[type="checkbox"]', { has: this.page.locator('label', { hasText: new RegExp(section, 'i') }) });
      await sectionCheckbox.check();
    }
    
    // Toggle charts and tables
    if (options.includeCharts) {
      const chartsToggle = this.page.locator('[class*="charts-toggle"], label:has-text("Charts")');
      await chartsToggle.click();
    }
    
    if (options.includeTables) {
      const tablesToggle = this.page.locator('[class*="tables-toggle"], label:has-text("Tables")');
      await tablesToggle.click();
    }
    
    const generateButton = this.page.locator('button:has-text("Generate Report"), [type="submit"]');
    await generateButton.click();
  }
}
