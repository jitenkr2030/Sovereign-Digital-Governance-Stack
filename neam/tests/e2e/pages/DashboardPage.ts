import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Dashboard page
 * Handles all dashboard-related interactions and verifications
 */
export class DashboardPage {
  readonly page: Page;
  
  // Navigation elements
  readonly sidebar: Locator;
  readonly navItems: Locator;
  readonly userMenu: Locator;
  readonly logoutButton: Locator;
  readonly userProfile: Locator;
  
  // Dashboard specific elements
  readonly welcomeMessage: Locator;
  readonly statsCards: Locator;
  readonly recentActivity: Locator;
  readonly alertsPanel: Locator;
  readonly quickActions: Locator;
  
  // Stats card elements
  readonly totalAlertsCard: Locator;
  readonly pendingReviewCard: Locator;
  readonly criticalCasesCard: Locator;
  readonly resolvedTodayCard: Locator;
  
  // Charts and visualizations
  readonly threatDistributionChart: Locator;
  readonly trendAnalysisChart: Locator;
  readonly geographicMap: Locator;
  
  // Search and filter
  readonly searchInput: Locator;
  readonly filterButton: Locator;
  readonly dateRangePicker: Locator;
  
  // Action buttons
  readonly createNewAlertButton: Locator;
  readonly generateReportButton: Locator;
  readonly refreshDataButton: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Navigation
    this.sidebar = page.locator('[data-testid="sidebar"], .sidebar, nav, [class*="sidebar"]');
    this.navItems = page.locator('[data-testid="nav-item"], .nav-item, nav a, [role="navigation"] a, [class*="nav-item"]');
    this.userMenu = page.locator('[data-testid="user-menu"], .user-menu, [class*="user-menu"], button:has-text("User"), button:has-text("Profile")');
    this.logoutButton = page.locator('[data-testid="logout-button"], .logout, a:has-text("Sign Out"), button:has-text("Logout"), button:has-text("Sign Out")');
    this.userProfile = page.locator('[data-testid="user-profile"], .user-profile, [class*="profile"]');
    
    // Dashboard content
    this.welcomeMessage = page.locator('[data-testid="welcome-message"], .welcome, h1:has-text("Welcome"), [class*="welcome"]');
    this.statsCards = page.locator('[data-testid="stats-card"], .stats-card, [class*="stats-card"]');
    this.recentActivity = page.locator('[data-testid="recent-activity"], .recent-activity, [class*="activity"], [class*="recent"]');
    this.alertsPanel = page.locator('[data-testid="alerts-panel"], .alerts-panel, [class*="alerts"], [role="region"][aria-label*="alert"]');
    this.quickActions = page.locator('[data-testid="quick-actions"], .quick-actions, [class*="quick-action"]');
    
    // Stats cards - specific selectors
    this.totalAlertsCard = page.locator('[data-testid="total-alerts-card"], [class*="total-alerts"]:has-text("Total Alerts")');
    this.pendingReviewCard = page.locator('[data-testid="pending-review-card"], [class*="pending-review"]:has-text("Pending Review")');
    this.criticalCasesCard = page.locator('[data-testid="critical-cases-card"], [class*="critical"]:has-text("Critical Cases")');
    this.resolvedTodayCard = page.locator('[data-testid="resolved-today-card"], [class*="resolved"]:has-text("Resolved Today")');
    
    // Charts
    this.threatDistributionChart = page.locator('[data-testid="threat-chart"], .threat-chart, [class*="chart"][class*="threat"]');
    this.trendAnalysisChart = page.locator('[data-testid="trend-chart"], .trend-chart, [class*="chart"][class*="trend"]');
    this.geographicMap = page.locator('[data-testid="geo-map"], .geo-map, [class*="map"][class*="geo"], [role="img"][aria-label*="map"]');
    
    // Controls
    this.searchInput = page.locator('[data-testid="search-input"], .search-input, input[type="search"], input[placeholder*="Search"]');
    this.filterButton = page.locator('[data-testid="filter-button"], .filter-button, button:has-text("Filter"), [class*="filter"]');
    this.dateRangePicker = page.locator('[data-testid="date-range-picker"], .date-picker, [class*="date-range"], input[type="date"]');
    this.createNewAlertButton = page.locator('[data-testid="create-alert-button"], button:has-text("Create Alert"), button:has-text("New Alert"), [class*="create"][class*="alert"]');
    this.generateReportButton = page.locator('[data-testid="generate-report-button"], button:has-text("Generate Report"), [class*="report"]');
    this.refreshDataButton = page.locator('[data-testid="refresh-button"], .refresh-button, button:has-text("Refresh"), [class*="refresh"]');
  }

  /**
   * Navigate to dashboard page
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/dashboard`, { waitUntil: 'networkidle' });
  }

  /**
   * Wait for dashboard to fully load
   */
  async waitForDashboardLoad(): Promise<void> {
    await expect(this.statsCards.first()).toBeVisible({ timeout: 10000 });
  }

  /**
   * Get the welcome message text
   */
  async getWelcomeMessage(): Promise<string> {
    const welcomeElement = this.welcomeMessage.first();
    await expect(welcomeElement).toBeVisible();
    return await welcomeElement.textContent() || '';
  }

  /**
   * Get count of visible stats cards
   */
  async getStatsCardCount(): Promise<number> {
    return await this.statsCards.count();
  }

  /**
   * Click on a navigation item by name
   * @param itemName - Name of the navigation item to click
   */
  async clickNavItem(itemName: string): Promise<void> {
    const item = this.navItems.filter({ hasText: new RegExp(itemName, 'i') });
    await expect(item.first()).toBeVisible();
    await item.first().click();
  }

  /**
   * Click on stats card by name
   * @param cardName - Name of the stats card (e.g., "Total Alerts", "Pending Review")
   */
  async clickStatsCard(cardName: string): Promise<void> {
    const card = this.page.locator('[class*="stats-card"]').filter({ hasText: new RegExp(cardName, 'i') });
    await expect(card.first()).toBeVisible();
    await card.first().click();
  }

  /**
   * Search for items
   * @param query - Search query
   */
  async search(query: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(query);
    await this.searchInput.press('Enter');
  }

  /**
   * Click filter button
   */
  async clickFilterButton(): Promise<void> {
    await this.filterButton.click();
  }

  /**
   * Click create new alert button
   */
  async clickCreateNewAlert(): Promise<void> {
    await this.createNewAlertButton.click();
  }

  /**
   * Click generate report button
   */
  async clickGenerateReport(): Promise<void> {
    await this.generateReportButton.click();
  }

  /**
   * Click refresh data button
   */
  async clickRefreshData(): Promise<void> {
    await this.refreshDataButton.click();
  }

  /**
   * Open user menu
   */
  async openUserMenu(): Promise<void> {
    await this.userMenu.click();
  }

  /**
   * Click logout button
   */
  async clickLogout(): Promise<void> {
    await this.openUserMenu();
    await this.logoutButton.click();
  }

  /**
   * Get number value from stats card
   * @param cardName - Name of the stats card
   */
  async getStatsCardValue(cardName: string): Promise<number> {
    const card = this.page.locator('[class*="stats-card"]').filter({ hasText: new RegExp(cardName, 'i') });
    await expect(card).toBeVisible();
    
    // Try to extract number from text content
    const text = await card.textContent() || '';
    const match = text.match(/\d+/);
    return match ? parseInt(match[0], 10) : 0;
  }

  /**
   * Verify dashboard stats are displayed
   */
  async verifyDashboardStats(): Promise<void> {
    await expect(this.totalAlertsCard).toBeVisible();
    await expect(this.pendingReviewCard).toBeVisible();
    await expect(this.criticalCasesCard).toBeVisible();
    await expect(this.resolvedTodayCard).toBeVisible();
  }

  /**
   * Verify dashboard charts are rendered
   */
  async verifyDashboardCharts(): Promise<void> {
    await expect(this.threatDistributionChart).toBeVisible();
    await expect(this.trendAnalysisChart).toBeVisible();
    await expect(this.geographicMap).toBeVisible();
  }

  /**
   * Get recent activity items count
   */
  async getRecentActivityCount(): Promise<number> {
    const activityItems = this.recentActivity.locator('[class*="activity-item"], [class*="item"]');
    return await activityItems.count();
  }

  /**
   * Get alerts count from alerts panel
   */
  async getAlertsCount(): Promise<number> {
    await expect(this.alertsPanel).toBeVisible();
    const alerts = this.alertsPanel.locator('[class*="alert-item"], [class*="alert"], tr, li');
    return await alerts.count();
  }

  /**
   * Check if sidebar is visible
   */
  async isSidebarVisible(): Promise<boolean> {
    return await this.sidebar.first().isVisible();
  }

  /**
   * Get all navigation item text
   */
  async getNavItems(): Promise<string[]> {
    const items = this.navItems.all();
    const textContents: string[] = [];
    for (const item of items) {
      if (await item.isVisible()) {
        const text = await item.textContent();
        if (text) textContents.push(text.trim());
      }
    }
    return textContents;
  }

  /**
   * Navigate to Intelligence page
   */
  async navigateToIntelligence(): Promise<void> {
    await this.clickNavItem('Intelligence');
  }

  /**
   * Navigate to Interventions page
   */
  async navigateToInterventions(): Promise<void> {
    await this.clickNavItem('Interventions');
  }

  /**
   * Navigate to Reports page
   */
  async navigateToReports(): Promise<void> {
    await this.clickNavItem('Reports');
  }

  /**
   * Navigate to Crisis Console
   */
  async navigateToCrisisConsole(): Promise<void> {
    await this.clickNavItem('Crisis Console');
  }

  /**
   * Wait for page to reload after data refresh
   */
  async waitForDataRefresh(): Promise<void> {
    // Wait for loading state to complete
    await this.page.waitForSelector('[class*="loading"]', { state: 'hidden' }).catch(() => {});
    
    // Wait for stats cards to be visible again
    await expect(this.statsCards.first()).toBeVisible({ timeout: 5000 });
  }

  /**
   * Get page title
   */
  async getPageTitle(): Promise<string> {
    return await this.page.title();
  }

  /**
   * Take screenshot of dashboard
   */
  async takeScreenshot(name: string): Promise<void> {
    await this.page.screenshot({ path: `test-results/screenshots/${name}-${Date.now()}.png` });
  }
}
