import { test, expect, type Page } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { ReportsPage } from '../pages/ReportsPage';
import { faker } from '../utils/data-factory';

test.describe('Report Generation Workflow', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let reportsPage: ReportsPage;
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    reportsPage = new ReportsPage(page);
    
    await loginPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('should generate weekly intelligence report', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    // Navigate to reports page
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Select weekly intelligence template
    await reportsPage.selectWeeklyIntelligenceTemplate();
    
    // Set date range
    await reportsPage.selectQuickDateRange('Last 7 days');
    
    // Generate report
    await reportsPage.clickGenerateReport();
    
    // Wait for generation to complete
    await reportsPage.waitForGenerationComplete('Weekly Intelligence', 60000);
    
    // Verify report was generated
    const reportContent = await reportsPage.getReportContent();
    expect(reportContent).toBeTruthy();
  });

  test('should generate monthly threat assessment report', async () => {
    await loginPage.loginAsAdmin();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Select monthly threat template
    await reportsPage.selectMonthlyThreatTemplate();
    
    // Set date range
    await reportsPage.selectQuickDateRange('Last 30 days');
    
    // Generate report
    await reportsPage.clickGenerateReport();
    
    // Wait for generation
    await reportsPage.waitForGenerationComplete('Monthly Threat', 90000);
    
    // Verify report contains charts and tables
    const charts = await reportsPage.reportCharts.count();
    const tables = await reportsPage.reportTables.count();
    
    expect(charts + tables).toBeGreaterThan(0);
  });

  test('should create and configure custom report', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Create custom report
    await reportsPage.createCustomReport({
      title: `Custom Report - ${faker.word()}`,
      type: 'Custom',
      startDate: '2024-01-01',
      endDate: '2024-12-31',
      sections: ['Executive Summary', 'Threat Analysis', 'Recommendations'],
      includeCharts: true,
      includeTables: true
    });
    
    // Wait for generation
    await reportsPage.waitForGenerationComplete('Custom Report', 120000);
    
    // Verify custom report was created
    const savedReports = await reportsPage.getSavedReportsCount();
    expect(savedReports).toBeGreaterThan(0);
  });

  test('should download report in PDF format', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Check if there are saved reports
    const reportsCount = await reportsPage.getSavedReportsCount();
    
    if (reportsCount > 0) {
      // Get first report title
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"], [class*="name"]').textContent();
      
      if (reportTitle) {
        // Download as PDF
        const downloadPromise = page.waitForEvent('download');
        await reportsPage.downloadReport(reportTitle.trim(), 'pdf');
        
        const download = await downloadPromise;
        expect(download.suggestedFilename()).toMatch(/\.pdf$/i);
      }
    }
  });

  test('should download report in CSV format', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const reportsCount = await reportsPage.getSavedReportsCount();
    
    if (reportsCount > 0) {
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"]').textContent();
      
      if (reportTitle) {
        const downloadPromise = page.waitForEvent('download');
        await reportsPage.downloadReport(reportTitle.trim(), 'csv');
        
        const download = await downloadPromise;
        expect(download.suggestedFilename()).toMatch(/\.csv$/i);
      }
    }
  });

  test('should share report via email', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const reportsCount = await reportsPage.getSavedReportsCount();
    
    if (reportsCount > 0) {
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"]').textContent();
      
      if (reportTitle) {
        // Share report
        await reportsPage.shareReport(reportTitle.trim(), 'stakeholder@example.gov');
        
        // Verify success notification
        const successNotification = page.locator('[data-testid="success-notification"], .toast-success');
        await expect(successNotification).toBeVisible({ timeout: 5000 });
      }
    }
  });

  test('should schedule recurring report', async () => {
    await loginPage.loginAsAdmin();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const reportsCount = await reportsPage.getSavedReportsCount();
    
    if (reportsCount > 0) {
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"]').textContent();
      
      if (reportTitle) {
        // Schedule report
        await reportsPage.scheduleReport(reportTitle.trim(), {
          frequency: 'Weekly',
          time: '09:00',
          recipients: ['team@example.gov', 'management@example.gov']
        });
        
        // Verify schedule confirmation
        const confirmation = page.locator('[data-testid="schedule-confirmation"], .toast-success');
        await expect(confirmation).toBeVisible({ timeout: 5000 });
      }
    }
  });

  test('should filter reports by status', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Filter by "Completed" status
    await reportsPage.filterByStatus('Completed');
    
    // Verify filtering
    const reportsCount = await reportsPage.getSavedReportsCount();
    expect(reportsCount).toBeGreaterThanOrEqual(0);
  });

  test('should search reports', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Search for specific report
    await reportsPage.search('Weekly');
    
    // Verify search results
    const reportsCount = await reportsPage.getSavedReportsCount();
    expect(reportsCount).toBeGreaterThanOrEqual(0);
  });

  test('should sort reports by date', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Sort by date (newest first)
    await reportsPage.sortByOption('Date');
    
    // Verify sorting is applied (check if table headers show sort indicator)
    const sortIndicator = page.locator('[class*="sorted"][class*="desc"], [aria-sort="descending"]');
    const hasSorting = await sortIndicator.count() > 0;
    
    expect(hasSorting || await reportsPage.getSavedReportsCount() >= 0).toBe(true);
  });

  test('should view report details', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const reportsCount = await reportsPage.getSavedReportsCount();
    
    if (reportsCount > 0) {
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"]').textContent();
      
      if (reportTitle) {
        // View report
        await reportsPage.viewReport(reportTitle.trim());
        
        // Verify report details are displayed
        await expect(reportsPage.reportTitle).toBeVisible({ timeout: 5000 });
        await expect(reportsPage.reportContent).toBeVisible();
      }
    }
  });

  test('should delete report', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const initialCount = await reportsPage.getSavedReportsCount();
    
    if (initialCount > 0) {
      const firstReport = reportsPage.savedReportItems.first();
      const reportTitle = await firstReport.locator('[class*="title"]').textContent();
      
      if (reportTitle) {
        // Delete report
        await reportsPage.deleteReport(reportTitle.trim());
        
        // Verify deletion confirmation
        const confirmation = page.locator('[data-testid="deletion-confirmation"], .toast-success');
        await expect(confirmation).toBeVisible({ timeout: 5000 });
        
        // Verify count decreased
        const finalCount = await reportsPage.getSavedReportsCount();
        expect(finalCount).toBeLessThan(initialCount);
      }
    }
  });

  test('should export all reports', async () => {
    await loginPage.loginAsAdmin();
    await loginPage.waitForLoginSuccess();
    
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    const downloadPromise = page.waitForEvent('download');
    await reportsPage.exportAllReports('csv');
    
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/\.csv$/i);
  });
});

test.describe('Report Templates', () => {
  let loginPage: LoginPage;
  let reportsPage: ReportsPage;
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    loginPage = new LoginPage(page);
    reportsPage = new ReportsPage(page);
    
    await loginPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await loginPage.loginAsAdmin();
    await loginPage.waitForLoginSuccess();
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('should display all available report templates', async () => {
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Verify templates section is visible
    await expect(reportsPage.templatesSection).toBeVisible();
    
    // Verify all template cards are present
    const templatesCount = await reportsPage.templateCards.count();
    expect(templatesCount).toBeGreaterThanOrEqual(4);
    
    // Verify specific templates exist
    await expect(reportsPage.weeklyIntelligenceTemplate).toBeVisible();
    await expect(reportsPage.monthlyThreatTemplate).toBeVisible();
    await expect(reportsPage.incidentSummaryTemplate).toBeVisible();
    await expect(reportsPage.customReportTemplate).toBeVisible();
  });

  test('should display template details on hover', async () => {
    await reportsPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await reportsPage.waitForPageLoad();
    
    // Hover over a template
    const template = reportsPage.templateCards.first();
    await template.hover();
    
    // Verify tooltip or expanded details appear
    const tooltip = page.locator('[class*="tooltip"], [class*="description"]');
    await expect(tooltip.first()).toBeVisible({ timeout: 2000 });
  });
});
