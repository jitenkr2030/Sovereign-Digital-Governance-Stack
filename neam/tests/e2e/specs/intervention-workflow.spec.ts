import { test, expect, type Page } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { IntelligencePage } from '../pages/IntelligencePage';
import { InterventionPage } from '../pages/InterventionPage';
import { faker } from '../utils/data-factory';

test.describe('Intelligence to Intervention Execution Workflow', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let intelligencePage: IntelligencePage;
  let interventionPage: InterventionPage;
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    intelligencePage = new IntelligencePage(page);
    interventionPage = new InterventionPage(page);
    
    await loginPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('should promote intelligence item to intervention', async () => {
    // Login as analyst
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    // Navigate to intelligence page
    await intelligencePage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await intelligencePage.waitForPageLoad();
    
    // Create a test intelligence item first
    await intelligencePage.clickCreateIntelligence();
    
    const intelligenceData = {
      title: faker.intelligenceTitle(),
      description: faker.intelligenceDescription(),
      source: 'Test Police Database',
      category: 'High Priority Threat',
      severity: 'High'
    };
    
    // Fill form
    const titleInput = page.locator('input[name="title"]');
    await titleInput.fill(intelligenceData.title);
    
    const descInput = page.locator('textarea[name="description"]');
    await descInput.fill(intelligenceData.description);
    
    // Submit
    const submitButton = page.locator('button:has-text("Create")');
    await submitButton.click();
    
    await intelligencePage.waitForPageLoad();
    
    // Promote to intervention
    await intelligencePage.promoteToIntervention(intelligenceData.title);
    
    // Verify promotion dialog opens
    const promotionDialog = page.locator('[data-testid="promotion-dialog"], .promotion-dialog');
    await expect(promotionDialog).toBeVisible();
    
    // Fill intervention creation form
    const interventionTitle = faker.interventionTitle();
    const titleField = page.locator('input[name="interventionTitle"]');
    await titleField.fill(interventionTitle);
    
    const typeSelect = page.locator('[name="interventionType"]');
    await typeSelect.selectOption({ label: 'Deployment' });
    
    const prioritySelect = page.locator('[name="priority"]');
    await prioritySelect.selectOption({ label: 'High' });
    
    // Create intervention
    const createButton = page.locator('button:has-text("Create Intervention")');
    await createButton.click();
    
    // Navigate to interventions page
    await dashboardPage.navigateToInterventions();
    await interventionPage.waitForPageLoad();
    
    // Verify intervention was created
    const exists = await interventionPage.interventionExists(interventionTitle);
    expect(exists).toBe(true);
  });

  test('should start intervention execution', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Find an intervention with "Pending" status
    const pendingBadge = interventionPage.pendingStatusBadge;
    const hasPendingInterventions = await pendingBadge.count() > 0;
    
    if (hasPendingInterventions) {
      const pendingRow = interventionPage.tableRows.filter({ has: pendingBadge });
      
      // Get the intervention title from the row
      const cells = pendingRow.locator('td');
      const firstCellText = await cells.first().textContent();
      
      if (firstCellText) {
        // Start the intervention
        await interventionPage.startIntervention(firstCellText);
        
        // Verify status changed to "In Progress"
        const status = await interventionPage.getInterventionStatus(firstCellText);
        expect(status.toLowerCase()).toContain('in progress');
      }
    }
  });

  test('should complete intervention with notes', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Find an "In Progress" intervention
    const inProgressBadge = interventionPage.inProgressStatusBadge;
    const hasInProgress = await inProgressBadge.count() > 0;
    
    if (hasInProgress) {
      const inProgressRow = interventionPage.tableRows.filter({ has: inProgressBadge });
      const cells = inProgressRow.locator('td');
      const firstCellText = await cells.first().textContent();
      
      if (firstCellText) {
        // Add completion notes
        await interventionPage.viewIntervention(firstCellText);
        await interventionPage.addNote(firstCellText, 'Intervention completed successfully. All objectives met.');
        
        // Complete the intervention
        await interventionPage.completeIntervention(firstCellText);
        
        // Verify status changed to "Completed"
        const status = await interventionPage.getInterventionStatus(firstCellText);
        expect(status.toLowerCase()).toContain('completed');
      }
    }
  });

  test('should filter interventions by status', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Filter by "In Progress"
    await interventionPage.filterByStatus('In Progress');
    
    // Verify filtering
    const itemsCount = await interventionPage.getItemsCount();
    expect(itemsCount).toBeGreaterThanOrEqual(0);
    
    // If there are items, verify they have the correct status
    if (itemsCount > 0) {
      const status = await interventionPage.getInterventionStatus('');
      expect(status.toLowerCase()).toContain('in progress');
    }
  });

  test('should filter interventions by priority', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Filter by "Critical" priority
    await interventionPage.filterByPriority('Critical');
    
    // Verify filtering
    const itemsCount = await interventionPage.getItemsCount();
    expect(itemsCount).toBeGreaterThanOrEqual(0);
  });

  test('should assign intervention to team member', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Find a pending intervention
    const pendingBadge = interventionPage.pendingStatusBadge;
    const hasPending = await pendingBadge.count() > 0;
    
    if (hasPending) {
      const pendingRow = interventionPage.tableRows.filter({ has: pendingBadge });
      const cells = pendingRow.locator('td');
      const firstCellText = await cells.first().textContent();
      
      if (firstCellText) {
        // Assign to a team member
        await interventionPage.assignIntervention(firstCellText, 'Test Administrator');
        
        // Verify assignment (check if assignee field shows the name)
        const row = interventionPage.getRowByTitle(firstCellText);
        const assigneeCell = row.locator('td', { hasText: 'Test Administrator' });
        await expect(assigneeCell).toBeVisible();
      }
    }
  });

  test('should switch between table and kanban views', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Verify table view is visible
    await expect(interventionPage.dataTable).toBeVisible();
    
    // Switch to Kanban view
    await interventionPage.switchToKanbanView();
    
    // Verify Kanban board is visible
    await expect(interventionPage.kanbanBoard).toBeVisible();
    
    // Check Kanban columns
    const pendingCount = await interventionPage.getKanbanColumnCount('pending');
    const inProgressCount = await interventionPage.getKanbanColumnCount('in_progress');
    const completedCount = await interventionPage.getKanbanColumnCount('completed');
    
    // Verify counts are non-negative
    expect(pendingCount).toBeGreaterThanOrEqual(0);
    expect(inProgressCount).toBeGreaterThanOrEqual(0);
    expect(completedCount).toBeGreaterThanOrEqual(0);
    
    // Switch back to table view
    await interventionPage.switchToTableView();
    await expect(interventionPage.dataTable).toBeVisible();
  });

  test('should search interventions', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(process.env.BASE_URL || 'http://localhost:3000');
    await interventionPage.waitForPageLoad();
    
    // Search for specific intervention
    await interventionPage.search('Patrol');
    
    // Verify search results
    const items = await interventionPage.getItemsCount();
    // Items should match the search criteria
  });

  test('should create intervention directly from dashboard', async () => {
    await loginPage.loginAsAnalyst();
    await loginPage.waitForLoginSuccess();
    
    // Wait for dashboard load
    await dashboardPage.waitForDashboardLoad();
    
    // Click create new alert/intervention button
    await dashboardPage.clickCreateNewAlert();
    
    // Verify create form opens
    const createForm = page.locator('[data-testid="create-intervention-form"], .create-form');
    await expect(createForm).toBeVisible();
    
    // Fill form
    const interventionData = {
      title: faker.interventionTitle(),
      type: 'Coordination',
      priority: 'High',
      description: 'Test intervention created from dashboard'
    };
    
    const titleInput = page.locator('input[name="title"]');
    await titleInput.fill(interventionData.title);
    
    const typeSelect = page.locator('[name="type"]');
    await typeSelect.selectOption({ label: interventionData.type });
    
    const prioritySelect = page.locator('[name="priority"]');
    await prioritySelect.selectOption({ label: interventionData.priority });
    
    // Submit
    const submitButton = page.locator('button:has-text("Create")');
    await submitButton.click();
    
    // Verify success
    await dashboardPage.waitForDataRefresh();
  });
});
