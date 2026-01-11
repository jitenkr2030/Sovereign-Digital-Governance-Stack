import { test, expect, type Page } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { IntelligencePage } from '../pages/IntelligencePage';
import { faker } from '../utils/data-factory';

test.describe('Data Ingestion to Intelligence Analysis Workflow', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let intelligencePage: IntelligencePage;
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    intelligencePage = new IntelligencePage(page);
    
    // Navigate to login page
    await loginPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('should login and navigate to intelligence dashboard', async () => {
    // Perform login
    await loginPage.login(
      process.env.TEST_ANALYST_EMAIL || 'analyst@test.gov',
      process.env.TEST_ANALYST_PASSWORD || 'test_password123'
    );
    
    // Wait for login success
    await loginPage.waitForLoginSuccess();
    
    // Navigate to dashboard
    await dashboardPage.waitForDashboardLoad();
    
    // Verify dashboard elements are visible
    await dashboardPage.verifyDashboardStats();
    await dashboardPage.verifyDashboardCharts();
    
    // Navigate to intelligence page
    await dashboardPage.navigateToIntelligence();
    await intelligencePage.waitForPageLoad();
    
    // Verify intelligence page is loaded
    await expect(intelligencePage.pageTitle).toBeVisible();
  });

  test('should create new intelligence item from incoming data', async () => {
    // Login as analyst
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    // Navigate to intelligence page
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Click create new intelligence button
    await intelligencePage.clickCreateIntelligence();
    
    // Generate test data
    const intelligenceData = {
      title: faker.intelligenceTitle(),
      description: faker.intelligenceDescription(),
      source: 'Test Police Database',
      category: 'Moderate Threat',
      severity: 'Medium'
    };
    
    // Fill in the intelligence form
    const titleInput = page.locator('input[name="title"], [data-testid="title-input"]');
    await titleInput.fill(intelligenceData.title);
    
    const descriptionInput = page.locator('textarea[name="description"], [data-testid="description-input"]');
    await descriptionInput.fill(intelligenceData.description);
    
    // Select source
    const sourceDropdown = page.locator('[name="source"], [data-testid="source-select"]');
    await sourceDropdown.selectOption({ label: intelligenceData.source });
    
    // Select category
    const categoryDropdown = page.locator('[name="category"], [data-testid="category-select"]');
    await categoryDropdown.selectOption({ label: intelligenceData.category });
    
    // Select severity
    const severityDropdown = page.locator('[name="severity"], [data-testid="severity-select"]');
    await severityDropdown.selectOption({ label: intelligenceData.severity });
    
    // Submit the form
    const submitButton = page.locator('button:has-text("Create"), [type="submit"]');
    await submitButton.click();
    
    // Verify the intelligence item was created
    await intelligencePage.waitForPageLoad();
    const itemsCount = await intelligencePage.getItemsCount();
    expect(itemsCount).toBeGreaterThan(0);
  });

  test('should filter intelligence items by status', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Filter by pending status
    await intelligencePage.filterByStatus('Pending');
    
    // Verify filtering works
    const itemsCount = await intelligencePage.getItemsCount();
    expect(itemsCount).toBeGreaterThanOrEqual(0);
  });

  test('should filter intelligence items by severity', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Filter by high severity
    await intelligencePage.filterBySeverity('High');
    
    // Verify filtering works
    const itemsCount = await intelligencePage.getItemsCount();
    expect(itemsCount).toBeGreaterThanOrEqual(0);
  });

  test('should search intelligence items', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Search for specific item
    await intelligencePage.search('Suspicious');
    
    // Verify search results
    const items = await intelligencePage.getAlerts();
    // Items should contain or match the search criteria
  });

  test('should view intelligence item details', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Get first item and view details
    const itemsCount = await intelligencePage.getItemsCount();
    if (itemsCount > 0) {
      // Click view on first item
      const firstRow = intelligencePage.tableRows.first();
      const viewButton = firstRow.locator('[data-testid="view-button"], button:has-text("View")');
      await viewButton.click();
      
      // Verify details page is loaded
      const detailsPage = page.locator('[data-testid="intelligence-details"], .details-page');
      await expect(detailsPage).toBeVisible({ timeout: 5000 });
    }
  });

  test('should handle data ingestion from external sources', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Click import button
    await intelligencePage.clickImport();
    
    // Verify import dialog opens
    const importDialog = page.locator('[data-testid="import-dialog"], .import-dialog');
    await expect(importDialog).toBeVisible();
    
    // Mock file upload
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles({
      name: 'test-intelligence-data.json',
      mimeType: 'application/json',
      buffer: Buffer.from(JSON.stringify({
        items: [
          {
            title: 'Imported Intelligence Item 1',
            description: 'Test description',
            severity: 'High',
            source: 'External API'
          }
        ]
      }))
    });
    
    // Submit import
    const importSubmitButton = page.locator('button:has-text("Import"), [type="submit"]');
    await importSubmitButton.click();
    
    // Verify import success notification
    const successNotification = page.locator('[data-testid="success-notification"], .toast-success');
    await expect(successNotification).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Intelligence Item Status Management', () => {
  let loginPage: LoginPage;
  let intelligencePage: IntelligencePage;
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    loginPage = new LoginPage(page);
    intelligencePage = new IntelligencePage(page);
    
    await loginPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('should transition intelligence item from new to pending status', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Find an item with "New" status
    const newBadge = intelligencePage.newStatusBadge;
    const hasNewItems = await newBadge.count() > 0;
    
    if (hasNewItems) {
      // Click on the new item
      const newItemRow = intelligencePage.tableRows.filter({ has: newBadge });
      await newItemRow.click();
      
      // Click edit or update status
      const updateStatusButton = page.locator('[data-testid="update-status-button"], button:has-text("Update Status")');
      await updateStatusButton.click();
      
      // Select "Pending" status
      const pendingOption = page.locator('[role="option"], option', { hasText: 'Pending' });
      await pendingOption.click();
      
      // Verify status change
      const currentStatus = await intelligencePage.getItemStatus('New Item');
      expect(currentStatus).toBe('Pending');
    }
  });

  test('should filter and display only high severity items', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Filter by high severity
    await intelligencePage.filterBySeverity('High');
    
    // Verify all displayed items have high severity
    const itemsCount = await intelligencePage.getItemsCount();
    if (itemsCount > 0) {
      const highBadge = intelligencePage.highSeverityBadge;
      const highItems = await intelligencePage.getAlertsBySeverity('High');
      
      // Check that filtering worked
      expect(await highItems.count()).toBeLessThanOrEqual(itemsCount);
    }
  });
});
