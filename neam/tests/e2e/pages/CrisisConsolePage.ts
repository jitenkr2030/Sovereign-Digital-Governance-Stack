import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Crisis Console page
 * Handles all crisis alert management and emergency response interactions
 */
export class CrisisConsolePage {
  readonly page: Page;
  
  // Page header and status
  readonly pageTitle: Locator;
  readonly consoleStatus: Locator;
  readonly activeAlertsCount: Locator;
  
  // Main console sections
  readonly alertFeed: Locator;
  readonly liveMap: Locator;
  readonly responseCoordination: Locator;
  readonly communicationPanel: Locator;
  
  // Alert actions
  readonly acknowledgeAlertButton: Locator;
  readonly escalateAlertButton: Locator;
  readonly resolveAlertButton: Locator;
  readonly createTaskButton: Locator;
  
  // Filter and view controls
  readonly severityFilter: Locator;
  readonly statusFilter: Locator;
  readonly viewToggle: Locator;
  readonly sortOrder: Locator;
  readonly autoRefreshToggle: Locator;
  
  // Alert list
  readonly alertList: Locator;
  readonly alertItems: Locator;
  readonly selectedAlert: Locator;
  readonly alertDetails: Locator;
  
  // Priority indicators
  readonly criticalAlertBadge: Locator;
  readonly highAlertBadge: Locator;
  readonly mediumAlertBadge: Locator;
  readonly lowAlertBadge: Locator;
  
  // Response actions
  readonly dispatchUnitButton: Locator;
  readonly activateProtocolButton: Locator;
  readonly notifyStakeholdersButton: Locator;
  readonly requestBackupButton: Locator;
  
  // Communication
  readonly emergencyBroadcastButton: Locator;
  readonly teamChatButton: Locator;
  readonly externalCommButton: Locator;
  readonly callLogButton: Locator;
  
  // Map controls
  readonly mapSearchInput: Locator;
  readonly mapLayersButton: Locator;
  readonly mapLegend: Locator;
  readonly mapZoomIn: Locator;
  readonly mapZoomOut: Locator;
  readonly centerOnAlertButton: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Page header
    this.pageTitle = page.locator('[data-testid="page-title"], h1:has-text("Crisis Console"), .page-title');
    this.consoleStatus = page.locator('[data-testid="console-status"], .console-status, [class*="status"]:has-text("Console")');
    this.activeAlertsCount = page.locator('[data-testid="active-alerts-count"], .active-alerts, [class*="alert-count"]');
    
    // Console sections
    this.alertFeed = page.locator('[data-testid="alert-feed"], .alert-feed, [class*="feed"]');
    this.liveMap = page.locator('[data-testid="live-map"], .live-map, [class*="map"]');
    this.responseCoordination = page.locator('[data-testid="coordination-panel"], .coordination-panel, [class*="coordination"]');
    this.communicationPanel = page.locator('[data-testid="communication-panel"], .comm-panel, [class*="communication"]');
    
    // Alert actions
    this.acknowledgeAlertButton = page.locator('[data-testid="acknowledge-button"], button:has-text("Acknowledge")');
    this.escalateAlertButton = page.locator('[data-testid="escalate-button"], button:has-text("Escalate")');
    this.resolveAlertButton = page.locator('[data-testid="resolve-button"], button:has-text("Resolve")');
    this.createTaskButton = page.locator('[data-testid="create-task-button"], button:has-text("Create Task"), button:has-text("New Task")');
    
    // Filters
    this.severityFilter = page.locator('[data-testid="severity-filter"], [class*="severity-filter"]');
    this.statusFilter = page.locator('[data-testid="status-filter"], [class*="status-filter"]');
    this.viewToggle = page.locator('[data-testid="view-toggle"], [class*="view-toggle"]');
    this.sortOrder = page.locator('[data-testid="sort-order"], [class*="sort-order"]');
    this.autoRefreshToggle = page.locator('[data-testid="auto-refresh"], [class*="auto-refresh"], input[type="checkbox"][name*="refresh"]');
    
    // Alert list
    this.alertList = page.locator('[data-testid="alert-list"], .alert-list, [class*="alert-list"]');
    this.alertItems = page.locator('[data-testid="alert-item"], .alert-item, [class*="alert-item"]');
    this.selectedAlert = page.locator('[data-testid="selected-alert"], .selected-alert, [class*="selected"][class*="alert"]');
    this.alertDetails = page.locator('[data-testid="alert-details"], .alert-details, [class*="alert-details"]');
    
    // Priority badges
    this.criticalAlertBadge = page.locator('[data-testid="severity-critical"], [class*="severity"]:has-text("Critical")');
    this.highAlertBadge = page.locator('[data-testid="severity-high"], [class*="severity"]:has-text("High")');
    this.mediumAlertBadge = page.locator('[data-testid="severity-medium"], [class*="severity"]:has-text("Medium")');
    this.lowAlertBadge = page.locator('[data-testid="severity-low"], [class*="severity"]:has-text("Low")');
    
    // Response actions
    this.dispatchUnitButton = page.locator('[data-testid="dispatch-button"], button:has-text("Dispatch Unit"), button:has-text("Dispatch")');
    this.activateProtocolButton = page.locator('[data-testid="activate-protocol-button"], button:has-text("Activate Protocol"), [class*="protocol"]');
    this.notifyStakeholdersButton = page.locator('[data-testid="notify-button"], button:has-text("Notify Stakeholders"), button:has-text("Notify")');
    this.requestBackupButton = page.locator('[data-testid="backup-button"], button:has-text("Request Backup"), button:has-text("Backup")');
    
    // Communication
    this.emergencyBroadcastButton = page.locator('[data-testid="broadcast-button"], button:has-text("Emergency Broadcast"), button:has-text("Broadcast")');
    this.teamChatButton = page.locator('[data-testid="team-chat-button"], button:has-text("Team Chat"), [class*="chat"]');
    this.externalCommButton = page.locator('[data-testid="external-comm-button"], button:has-text("External Comms"), [class*="external"]');
    this.callLogButton = page.locator('[data-testid="call-log-button"], button:has-text("Call Log"), [class*="call"]');
    
    // Map controls
    this.mapSearchInput = page.locator('[data-testid="map-search"], input[placeholder*="Search map"], [class*="map-search"]');
    this.mapLayersButton = page.locator('[data-testid="map-layers"], button:has-text("Layers"), [class*="layers"]');
    this.mapLegend = page.locator('[data-testid="map-legend"], .map-legend, [class*="legend"]');
    this.mapZoomIn = page.locator('[data-testid="zoom-in"], button:has-text("Zoom In"), [class*="zoom-in"]');
    this.mapZoomOut = page.locator('[data-testid="zoom-out"], button:has-text("Zoom Out"), [class*="zoom-out"]');
    this.centerOnAlertButton = page.locator('[data-testid="center-alert"], button:has-text("Center on Alert"), [class*="center"]');
  }

  /**
   * Navigate to crisis console
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/crisis-console`, { waitUntil: 'networkidle' });
  }

  /**
   * Wait for console to fully load
   */
  async waitForConsoleLoad(): Promise<void> {
    await expect(this.alertFeed.or(this.liveMap)).toBeVisible({ timeout: 10000 });
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
   * Get active alerts count
   */
  async getActiveAlertsCount(): Promise<number> {
    const countText = await this.activeAlertsCount.textContent();
    const match = countText?.match(/\d+/);
    return match ? parseInt(match[0], 10) : 0;
  }

  /**
   * Get list of all alert items
   */
  async getAlerts(): Promise<string[]> {
    await expect(this.alertItems.first()).toBeVisible();
    const texts = await this.alertItems.allTextContents();
    return texts.map(t => t.trim());
  }

  /**
   * Get alerts by severity
   * @param severity - Severity level (critical, high, medium, low)
   */
  async getAlertsBySeverity(severity: string): Promise<Locator> {
    const badge = this.getSeverityBadge(severity);
    return this.alertItems.filter({ has: badge });
  }

  /**
   * Get severity badge locator
   * @param severity - Severity level
   */
  private getSeverityBadge(severity: string): Locator {
    const normalizedSeverity = severity.toLowerCase();
    switch (normalizedSeverity) {
      case 'critical':
        return this.criticalAlertBadge;
      case 'high':
        return this.highAlertBadge;
      case 'medium':
        return this.mediumAlertBadge;
      case 'low':
        return this.lowAlertBadge;
      default:
        throw new Error(`Unknown severity: ${severity}`);
    }
  }

  /**
   * Select an alert from the list
   * @param alertTitle - Title or identifier of the alert
   */
  async selectAlert(alertTitle: string): Promise<void> {
    const alertItem = this.alertItems.filter({ hasText: new RegExp(alertTitle, 'i') });
    await expect(alertItem).toBeVisible();
    await alertItem.click();
  }

  /**
   * Acknowledge selected alert
   */
  async acknowledgeSelectedAlert(): Promise<void> {
    await this.acknowledgeAlertButton.click();
  }

  /**
   * Acknowledge specific alert
   * @param alertTitle - Title of the alert
   */
  async acknowledgeAlert(alertTitle: string): Promise<void> {
    await this.selectAlert(alertTitle);
    await this.acknowledgeSelectedAlert();
  }

  /**
   * Escalate selected alert
   */
  async escalateSelectedAlert(): Promise<void> {
    await this.escalateAlertButton.click();
  }

  /**
   * Escalate specific alert
   * @param alertTitle - Title of the alert
   */
  async escalateAlert(alertTitle: string): Promise<void> {
    await this.selectAlert(alertTitle);
    await this.escalateSelectedAlert();
  }

  /**
   * Resolve selected alert
   */
  async resolveSelectedAlert(): Promise<void> {
    await this.resolveAlertButton.click();
  }

  /**
   * Resolve specific alert
   * @param alertTitle - Title of the alert
   */
  async resolveAlert(alertTitle: string): Promise<void> {
    await this.selectAlert(alertTitle);
    await this.resolveSelectedAlert();
  }

  /**
   * Create task for selected alert
   * @param taskTitle - Title of the task
   * @param description - Task description
   */
  async createTaskForAlert(taskTitle: string, description?: string): Promise<void> {
    await this.createTaskButton.click();
    // Fill in task details in modal/dialog
    const titleInput = this.page.locator('input[name="taskTitle"], input[placeholder*="Title"]');
    await titleInput.fill(taskTitle);
    
    if (description) {
      const descInput = this.page.locator('textarea[name="description"], [contenteditable]');
      await descInput.fill(description);
    }
    
    const submitButton = this.page.locator('button:has-text("Create"), button:has-text("Save"), [type="submit"]');
    await submitButton.click();
  }

  /**
   * Filter by severity
   * @param severity - Severity level
   */
  async filterBySeverity(severity: string): Promise<void> {
    await this.severityFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(severity, 'i') });
    await option.click();
  }

  /**
   * Filter by status
   * @param status - Status (new, acknowledged, escalated, resolved)
   */
  async filterByStatus(status: string): Promise<void> {
    await this.statusFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(status, 'i') });
    await option.click();
  }

  /**
   * Toggle auto-refresh
   */
  async toggleAutoRefresh(): Promise<void> {
    await this.autoRefreshToggle.click();
  }

  /**
   * Toggle between views (list/map)
   */
  async toggleView(view: 'list' | 'map' | 'split'): Promise<void> {
    await this.viewToggle.click();
    const viewOption = this.page.locator('[role="option"], button', { hasText: new RegExp(view, 'i') });
    await viewOption.click();
  }

  /**
   * Dispatch unit to alert location
   * @param alertTitle - Title of the alert
   * @param unitName - Name of the unit to dispatch
   */
  async dispatchUnit(alertTitle: string, unitName: string): Promise<void> {
    await this.selectAlert(alertTitle);
    await this.dispatchUnitButton.click();
    
    // Select unit from dropdown
    const unitOption = this.page.locator('[role="listbox"]', { hasText: new RegExp(unitName, 'i') });
    await unitOption.click();
  }

  /**
   * Activate emergency protocol
   * @param protocolName - Name of the protocol to activate
   */
  async activateProtocol(protocolName: string): Promise<void> {
    await this.activateProtocolButton.click();
    const protocolOption = this.page.locator('[role="option"]', { hasText: new RegExp(protocolName, 'i') });
    await protocolOption.click();
  }

  /**
   * Send emergency broadcast
   * @param message - Broadcast message
   */
  async sendEmergencyBroadcast(message: string): Promise<void> {
    await this.emergencyBroadcastButton.click();
    const messageInput = this.page.locator('textarea, [contenteditable]');
    await messageInput.fill(message);
    
    const sendButton = this.page.locator('button:has-text("Send"), button:has-text("Broadcast")');
    await sendButton.click();
  }

  /**
   * Get alert details
   * @param alertTitle - Title of the alert
   */
  async getAlertDetails(alertTitle: string): Promise<string> {
    await this.selectAlert(alertTitle);
    await expect(this.alertDetails).toBeVisible();
    return await this.alertDetails.textContent() || '';
  }

  /**
   * Check if critical alerts exist
   */
  async hasCriticalAlerts(): Promise<boolean> {
    const criticalAlerts = await this.getAlertsBySeverity('critical');
    return await criticalAlerts.count() > 0;
  }

  /**
   * Wait for new alerts
   */
  async waitForNewAlerts(previousCount: number, timeout: number = 30000): Promise<void> {
    await this.page.waitForFunction(
      (previousCount: number) => {
        const alertItems = document.querySelectorAll('[class*="alert-item"], .alert-item');
        return alertItems.length > previousCount;
      },
      previousCount,
      { timeout }
    );
  }

  /**
   * Get map zoom level
   */
  async getMapZoomLevel(): Promise<number> {
    // This depends on the map library implementation
    const mapContainer = this.liveMap.locator('[class*="leaflet"]');
    if (await mapContainer.isVisible()) {
      // Leaflet map example
      const zoomButton = this.page.locator('.leaflet-control-zoom a');
      const zoomInLevel = await zoomButton.first().getAttribute('data-zoom');
      return parseInt(zoomInLevel || '10', 10);
    }
    return 10;
  }

  /**
   * Center map on specific alert
   * @param alertTitle - Title of the alert
   */
  async centerMapOnAlert(alertTitle: string): Promise<void> {
    await this.selectAlert(alertTitle);
    await this.centerOnAlertButton.click();
  }

  /**
   * Open team chat
   */
  async openTeamChat(): Promise<void> {
    await this.teamChatButton.click();
  }

  /**
   * Get console status
   */
  async getConsoleStatus(): Promise<string> {
    return await this.consoleStatus.textContent() || '';
  }
}
