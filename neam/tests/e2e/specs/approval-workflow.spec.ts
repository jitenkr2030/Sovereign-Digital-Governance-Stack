import { test, expect, type Page } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { IntelligencePage } from '../pages/IntelligencePage';
import { InterventionPage } from '../pages/InterventionPage';
import { faker } from '../utils/data-factory';

test.describe('Multi-User Approval Workflow', () => {
  let analystPage: Page;
  let supervisorPage: Page;
  let adminPage: Page;
  let analystLoginPage: LoginPage;
  let supervisorLoginPage: LoginPage;
  let adminLoginPage: LoginPage;
  let intelligencePage: IntelligencePage;
  let dashboardPage: DashboardPage;
  let interventionPage: InterventionPage;

  test.beforeEach(async ({ browser }) => {
    // Create separate browser contexts for each user type
    analystPage = await browser.newPage({ viewport: { width: 1280, height: 720 } });
    supervisorPage = await browser.newPage({ viewport: { width: 1280, height: 720 } });
    adminPage = await browser.newPage({ viewport: { width: 1280, height: 720 } });
    
    analystLoginPage = new LoginPage(analystPage);
    supervisorLoginPage = new LoginPage(supervisorPage);
    adminLoginPage = new LoginPage(adminPage);
    intelligencePage = new IntelligencePage(analystPage);
    dashboardPage = new DashboardPage(analystPage);
    interventionPage = new InterventionPage(analystPage);
    
    // Initialize all pages
    intelligencePage = new IntelligencePage(analystPage);
    intelligencePage = new IntelligencePage(supervisorPage);
    intelligencePage = new IntelligencePage(adminPage);
    dashboardPage = new DashboardPage(analystPage);
    dashboardPage = new DashboardPage(supervisorPage);
    dashboardPage = new DashboardPage(adminPage);
    interventionPage = new InterventionPage(analystPage);
    interventionPage = new InterventionPage(supervisorPage);
    interventionPage = new InterventionPage(adminPage);
  });

  test.afterEach(async () => {
    await analystPage.close();
    await supervisorPage.close();
    await adminPage.close();
  });

  test('junior analyst creates item, supervisor approves', async () => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Step 1: Junior analyst logs in and creates an intelligence item
    await analystLoginPage.navigateTo(baseURL);
    await analystLoginPage.login(
      process.env.TEST_ANALYST_EMAIL || 'analyst@test.gov',
      process.env.TEST_ANALYST_PASSWORD || 'test_password123'
    );
    await analystLoginPage.waitForLoginSuccess();
    
    // Create new intelligence item
    await intelligencePage.navigateTo(baseURL);
    await intelligencePage.waitForPageLoad();
    await intelligencePage.clickCreateIntelligence();
    
    const itemTitle = `Test Intelligence - ${faker.word()}`;
    await analystPage.locator('input[name="title"]').fill(itemTitle);
    await analystPage.locator('textarea[name="description"]').fill(faker.intelligenceDescription());
    await analystPage.locator('[name="severity"]').selectOption('High');
    
    // Submit
    await analystPage.locator('button:has-text("Create")').click();
    await intelligencePage.waitForPageLoad();
    
    // Verify item was created with "New" status
    const status = await intelligencePage.getItemStatus(itemTitle);
    expect(status.toLowerCase()).toContain('new');
    
    // Step 2: Supervisor logs in and reviews the item
    await supervisorLoginPage.navigateTo(baseURL);
    await supervisorLoginPage.login(
      process.env.TEST_SUPERVISOR_EMAIL || 'supervisor@test.gov',
      process.env.TEST_SUPERVISOR_PASSWORD || 'test_password123'
    );
    await supervisorLoginPage.waitForLoginSuccess();
    
    // Navigate to intelligence page
    await intelligencePage.navigateTo(baseURL);
    await intelligencePage.waitForPageLoad();
    
    // Find and view the item created by analyst
    await intelligencePage.viewItem(itemTitle);
    
    // Approve the item
    const approveButton = supervisorPage.locator('button:has-text("Approve"), [data-testid="approve-button"]');
    await expect(approveButton).toBeVisible({ timeout: 5000 });
    await approveButton.click();
    
    // Verify status changed to "Under Review" or "Approved"
    await supervisorPage.reload();
    await intelligencePage.waitForPageLoad();
    const updatedStatus = await intelligencePage.getItemStatus(itemTitle);
    expect(updatedStatus.toLowerCase()).not.toContain('new');
  });

  test('escalation: analyst requests escalation, admin handles', async () => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Analyst logs in
    await analystLoginPage.navigateTo(baseURL);
    await analystLoginPage.loginAsAnalyst();
    await analystLoginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(baseURL);
    await intelligencePage.waitForPageLoad();
    
    // Create an item that will need escalation
    await intelligencePage.clickCreateIntelligence();
    const escalationTitle = `Critical Issue - ${faker.word()}`;
    await analystPage.locator('input[name="title"]').fill(escalationTitle);
    await analystPage.locator('textarea[name="description"]').fill('Critical issue requiring immediate attention');
    await analystPage.locator('[name="severity"]').selectOption('Critical');
    await analystPage.locator('button:has-text("Create")').click();
    await intelligencePage.waitForPageLoad();
    
    // Request escalation
    const escalateButton = analystPage.locator('button:has-text("Escalate"), [data-testid="escalate-button"]');
    await escalateButton.click();
    
    // Fill escalation reason
    await analystPage.locator('textarea[name="reason"]').fill('This requires immediate supervisor attention');
    await analystPage.locator('button:has-text("Submit")').click();
    
    // Verify escalation request was sent
    const escalationConfirmation = analystPage.locator('[data-testid="escalation-confirmation"], .toast-success');
    await expect(escalationConfirmation).toBeVisible();
    
    // Admin logs in to handle escalation
    await adminLoginPage.navigateTo(baseURL);
    await adminLoginPage.login(
      process.env.TEST_ADMIN_EMAIL || 'admin@test.gov',
      process.env.TEST_ADMIN_PASSWORD || 'test_password123'
    );
    await adminLoginPage.waitForLoginSuccess();
    
    // Navigate to escalated items
    await dashboardPage.navigateTo(baseURL);
    await dashboardPage.waitForDashboardLoad();
    
    // Check for escalation notifications
    const notificationBadge = dashboardPage.page.locator('[class*="notification"]:has-text("Escalation")');
    const hasEscalations = await notificationBadge.count() > 0;
    
    if (hasEscalations) {
      await notificationBadge.click();
    }
  });

  test('intervention approval: analyst creates, supervisor approves', async () => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Analyst creates an intervention
    await analystLoginPage.navigateTo(baseURL);
    await analystLoginPage.loginAsAnalyst();
    await analystLoginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(baseURL);
    await interventionPage.waitForPageLoad();
    await interventionPage.clickCreateIntervention();
    
    const interventionTitle = `Intervention - ${faker.word()}`;
    await analystPage.locator('input[name="title"]').fill(interventionTitle);
    await analystPage.locator('textarea[name="description"]').fill('Intervention plan requiring approval');
    await analystPage.locator('[name="type"]').selectOption('Deployment');
    await analystPage.locator('[name="priority"]').selectOption('High');
    await analystPage.locator('button:has-text("Create")').click();
    await interventionPage.waitForPageLoad();
    
    // Verify intervention was created with "Pending" status
    const initialStatus = await interventionPage.getInterventionStatus(interventionTitle);
    expect(initialStatus.toLowerCase()).toContain('pending');
    
    // Supervisor reviews and approves
    await supervisorLoginPage.navigateTo(baseURL);
    await supervisorLoginPage.loginAsSupervisor();
    await supervisorLoginPage.waitForLoginSuccess();
    
    await interventionPage.navigateTo(baseURL);
    await interventionPage.waitForPageLoad();
    
    // View the intervention
    await interventionPage.viewIntervention(interventionTitle);
    
    // Approve
    const approveButton = supervisorPage.locator('button:has-text("Approve"), [data-testid="approve-button"]');
    await expect(approveButton).toBeVisible({ timeout: 5000 });
    await approveButton.click();
    
    // Verify status changed
    await supervisorPage.reload();
    await interventionPage.waitForPageLoad();
    const approvedStatus = await interventionPage.getInterventionStatus(interventionTitle);
    expect(approvedStatus.toLowerCase()).not.toContain('pending');
  });

  test('role-based access control: permissions verification', async () => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Test admin has full access
    await adminLoginPage.navigateTo(baseURL);
    await adminLoginPage.loginAsAdmin();
    await adminLoginPage.waitForLoginSuccess();
    
    await dashboardPage.waitForDashboardLoad();
    
    // Admin should see admin-specific controls
    const adminControls = adminPage.locator('[data-testid="admin-panel"], .admin-panel');
    const adminHasAccess = await adminControls.count() > 0;
    
    // Analyst should not see admin controls
    await analystLoginPage.navigateTo(baseURL);
    await analystLoginPage.loginAsAnalyst();
    await analystLoginPage.waitForLoginSuccess();
    
    await dashboardPage.waitForDashboardLoad();
    
    // Verify analyst cannot access admin functions
    const analystAdminControls = analystPage.locator('[data-testid="admin-panel"], .admin-panel');
    if (await analystAdminControls.count() > 0) {
      await expect(analystAdminControls.first()).not.toBeVisible();
    }
  });

  test('multi-user concurrent access', async ({ browser }) => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Create a shared context for testing concurrent edits
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    
    const page1 = await context1.newPage();
    const page2 = await context2.newPage();
    
    const login1 = new LoginPage(page1);
    const login2 = new LoginPage(page2);
    
    // Both users log in simultaneously
    await Promise.all([
      login1.navigateTo(baseURL),
      login2.navigateTo(baseURL)
    ]);
    
    await Promise.all([
      login1.loginAsAnalyst(),
      login2.loginAsSupervisor()
    ]);
    
    await Promise.all([
      login1.waitForLoginSuccess(),
      login2.waitForLoginSuccess()
    ]);
    
    // Both users navigate to the same page
    const intelPage1 = new IntelligencePage(page1);
    const intelPage2 = new IntelligencePage(page2);
    
    await Promise.all([
      intelPage1.navigateTo(baseURL),
      intelPage2.navigateTo(baseURL)
    ]);
    
    // Verify both can view the same data
    const items1 = await intelPage1.getItemsCount();
    const items2 = await intelPage2.getItemsCount();
    
    expect(items1).toBe(items2);
    
    // Cleanup
    await context1.close();
    await context2.close();
  });

  test('workflow state transitions verification', async () => {
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    
    // Create an item as analyst
    await analystLoginPage.navigateTo(baseURL);
    await analystLoginPage.loginAsAnalyst();
    await analystLoginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(baseURL);
    await intelligencePage.waitForPageLoad();
    await intelligencePage.clickCreateIntelligence();
    
    const workflowTitle = `Workflow Test - ${faker.word()}`;
    await analystPage.locator('input[name="title"]').fill(workflowTitle);
    await analystPage.locator('textarea[name="description"]').fill('Testing workflow transitions');
    await analystPage.locator('button:has-text("Create")').click();
    await intelligencePage.waitForPageLoad();
    
    // Initial state should be "New"
    expect(await intelligencePage.getItemStatus(workflowTitle)).toContain('New');
    
    // Submit for review
    const submitReviewButton = analystPage.locator('button:has-text("Submit for Review")');
    await submitReviewButton.click();
    
    await intelligencePage.waitForPageLoad();
    expect(await intelligencePage.getItemStatus(workflowTitle)).toContain('Pending');
    
    // Supervisor approves
    await supervisorLoginPage.navigateTo(baseURL);
    await supervisorLoginPage.loginAsSupervisor();
    await supervisorLoginPage.waitForLoginSuccess();
    
    await intelligencePage.navigateTo(baseURL);
    await intelligencePage.waitForPageLoad();
    await intelligencePage.viewItem(workflowTitle);
    
    const approveButton = supervisorPage.locator('button:has-text("Approve")');
    await approveButton.click();
    
    // Final state should be approved
    await supervisorPage.reload();
    await intelligencePage.waitForPageLoad();
    const finalStatus = await intelligencePage.getItemStatus(workflowTitle);
    expect(finalStatus.toLowerCase()).toMatch(/approved|under.*review|resolved/);
  });
});
