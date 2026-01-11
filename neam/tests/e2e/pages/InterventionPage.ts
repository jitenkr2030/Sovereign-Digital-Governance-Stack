import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Intervention page
 * Handles all intervention management interactions and verifications
 */
export class InterventionPage {
  readonly page: Page;
  
  // Page header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;
  
  // Action buttons
  readonly createInterventionButton: Locator;
  readonly assignButton: Locator;
  readonly exportButton: Locator;
  readonly refreshButton: Locator;
  
  // Search and filters
  readonly searchInput: Locator;
  readonly statusFilter: Locator;
  readonly priorityFilter: Locator;
  readonly typeFilter: Locator;
  readonly assigneeFilter: Locator;
  readonly clearFiltersButton: Locator;
  
  // Data table
  readonly dataTable: Locator;
  readonly tableHeaders: Locator;
  readonly tableRows: Locator;
  
  // Kanban board (alternative view)
  readonly kanbanBoard: Locator;
  readonly pendingColumn: Locator;
  readonly inProgressColumn: Locator;
  readonly completedColumn: Locator;
  readonly blockedColumn: Locator;
  
  // Individual intervention actions
  readonly viewInterventionButton: Locator;
  readonly editInterventionButton: Locator;
  readonly startInterventionButton: Locator;
  readonly completeInterventionButton: Locator;
  readonly assignInterventionButton: Locator;
  
  // Status indicators
  readonly pendingStatusBadge: Locator;
  readonly inProgressStatusBadge: Locator;
  readonly completedStatusBadge: Locator;
  readonly blockedStatusBadge: Locator;
  
  // Priority indicators
  readonly lowPriorityBadge: Locator;
  readonly mediumPriorityBadge: Locator;
  readonly highPriorityBadge: Locator;
  readonly criticalPriorityBadge: Locator;
  
  // Timeline/Activity
  readonly activityTimeline: Locator;
  readonly addNoteButton: Locator;
  readonly updateStatusButton: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Page header
    this.pageTitle = page.locator('[data-testid="page-title"], h1:has-text("Intervention"), .page-title');
    this.pageDescription = page.locator('[data-testid="page-description"], .page-description');
    
    // Actions
    this.createInterventionButton = page.locator('[data-testid="create-intervention-button"], button:has-text("Create Intervention"), button:has-text("New Intervention"), [class*="create"][class*="intervention"]');
    this.assignButton = page.locator('[data-testid="assign-button"], button:has-text("Assign"), [class*="assign"]');
    this.exportButton = page.locator('[data-testid="export-button"], button:has-text("Export"), [class*="export"]');
    this.refreshButton = page.locator('[data-testid="refresh-button"], button:has-text("Refresh"), [class*="refresh"]');
    
    // Filters
    this.searchInput = page.locator('[data-testid="search-input"], input[type="search"]');
    this.statusFilter = page.locator('[data-testid="status-filter"], [class*="status-filter"]');
    this.priorityFilter = page.locator('[data-testid="priority-filter"], [class*="priority-filter"]');
    this.typeFilter = page.locator('[data-testid="type-filter"], [class*="type-filter"]');
    this.assigneeFilter = page.locator('[data-testid="assignee-filter"], [class*="assignee-filter"]');
    this.clearFiltersButton = page.locator('[data-testid="clear-filters-button"], button:has-text("Clear Filters")');
    
    // Data table
    this.dataTable = page.locator('[data-testid="intervention-table"], table, [class*="data-table"]');
    this.tableHeaders = page.locator('[data-testid="intervention-table"] th, table th');
    this.tableRows = page.locator('[data-testid="intervention-table"] tbody tr, table tbody tr');
    
    // Kanban board
    this.kanbanBoard = page.locator('[data-testid="kanban-board"], .kanban-board, [class*="kanban"]');
    this.pendingColumn = page.locator('[data-testid="pending-column"], [class*="pending"][class*="column"]');
    this.inProgressColumn = page.locator('[data-testid="in-progress-column"], [class*="in-progress"][class*="column"]');
    this.completedColumn = page.locator('[data-testid="completed-column"], [class*="completed"][class*="column"]');
    this.blockedColumn = page.locator('[data-testid="blocked-column"], [class*="blocked"][class*="column"]');
    
    // Item actions
    this.viewInterventionButton = page.locator('[data-testid="view-button"], button:has-text("View")');
    this.editInterventionButton = page.locator('[data-testid="edit-button"], button:has-text("Edit")');
    this.startInterventionButton = page.locator('[data-testid="start-button"], button:has-text("Start"), button:has-text("Begin")');
    this.completeInterventionButton = page.locator('[data-testid="complete-button"], button:has-text("Complete")');
    this.assignInterventionButton = page.locator('[data-testid="assign-button"], button:has-text("Assign")');
    
    // Status badges
    this.pendingStatusBadge = page.locator('[data-testid="status-pending"], [class*="status"]:has-text("Pending")');
    this.inProgressStatusBadge = page.locator('[data-testid="status-in-progress"], [class*="status"]:has-text("In Progress")');
    this.completedStatusBadge = page.locator('[data-testid="status-completed"], [class*="status"]:has-text("Completed")');
    this.blockedStatusBadge = page.locator('[data-testid="status-blocked"], [class*="status"]:has-text("Blocked")');
    
    // Priority badges
    this.lowPriorityBadge = page.locator('[data-testid="priority-low"], [class*="priority"]:has-text("Low")');
    this.mediumPriorityBadge = page.locator('[data-testid="priority-medium"], [class*="priority"]:has-text("Medium")');
    this.highPriorityBadge = page.locator('[data-testid="priority-high"], [class*="priority"]:has-text("High")');
    this.criticalPriorityBadge = page.locator('[data-testid="priority-critical"], [class*="priority"]:has-text("Critical")');
    
    // Activity
    this.activityTimeline = page.locator('[data-testid="activity-timeline"], .activity-timeline, [class*="timeline"]');
    this.addNoteButton = page.locator('[data-testid="add-note-button"], button:has-text("Add Note"), button:has-text("Add Comment")');
    this.updateStatusButton = page.locator('[data-testid="update-status-button"], button:has-text("Update Status")');
  }

  /**
   * Navigate to interventions page
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/interventions`, { waitUntil: 'networkidle' });
  }

  /**
   * Wait for page to fully load
   */
  async waitForPageLoad(): Promise<void> {
    await expect(this.dataTable.or(this.kanbanBoard)).toBeVisible({ timeout: 10000 });
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
   * Get total count of interventions
   */
  async getItemsCount(): Promise<number> {
    return await this.tableRows.count();
  }

  /**
   * Search for interventions
   * @param query - Search query
   */
  async search(query: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(query);
    await this.searchInput.press('Enter');
    await this.waitForPageLoad();
  }

  /**
   * Filter by status
   * @param status - Status to filter by
   */
  async filterByStatus(status: string): Promise<void> {
    await this.statusFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(status, 'i') });
    await option.click();
    await this.waitForPageLoad();
  }

  /**
   * Filter by priority
   * @param priority - Priority level to filter by
   */
  async filterByPriority(priority: string): Promise<void> {
    await this.priorityFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(priority, 'i') });
    await option.click();
    await this.waitForPageLoad();
  }

  /**
   * Filter by type
   * @param type - Intervention type to filter by
   */
  async filterByType(type: string): Promise<void> {
    await this.typeFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(type, 'i') });
    await option.click();
    await this.waitForPageLoad();
  }

  /**
   * Filter by assignee
   * @param assignee - Assignee name to filter by
   */
  async filterByAssignee(assignee: string): Promise<void> {
    await this.assigneeFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(assignee, 'i') });
    await option.click();
    await this.waitForPageLoad();
  }

  /**
   * Clear all filters
   */
  async clearFilters(): Promise<void> {
    await this.clearFiltersButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Click create new intervention button
   */
  async clickCreateIntervention(): Promise<void> {
    await this.createInterventionButton.click();
  }

  /**
   * Click refresh button
   */
  async clickRefresh(): Promise<void> {
    await this.refreshButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Get row by intervention title
   * @param interventionTitle - Title to search for
   */
  private getRowByTitle(interventionTitle: string): Locator {
    return this.tableRows.filter({ has: this.page.locator('td', { hasText: new RegExp(interventionTitle, 'i') }) });
  }

  /**
   * View intervention details
   * @param interventionTitle - Title of the intervention
   */
  async viewIntervention(interventionTitle: string): Promise<void> {
    const row = this.getRowByTitle(interventionTitle);
    const viewButton = row.locator(this.viewInterventionButton);
    await expect(viewButton).toBeVisible();
    await viewButton.click();
  }

  /**
   * Edit intervention
   * @param interventionTitle - Title of the intervention
   */
  async editIntervention(interventionTitle: string): Promise<void> {
    const row = this.getRowByTitle(interventionTitle);
    const editButton = row.locator(this.editInterventionButton);
    await expect(editButton).toBeVisible();
    await editButton.click();
  }

  /**
   * Start intervention
   * @param interventionTitle - Title of the intervention
   */
  async startIntervention(interventionTitle: string): Promise<void> {
    const row = this.getRowByTitle(interventionTitle);
    const startButton = row.locator(this.startInterventionButton);
    await expect(startButton).toBeVisible();
    await startButton.click();
  }

  /**
   * Complete intervention
   * @param interventionTitle - Title of the intervention
   */
  async completeIntervention(interventionTitle: string): Promise<void> {
    const row = this.getRowByTitle(interventionTitle);
    const completeButton = row.locator(this.completeInterventionButton);
    await expect(completeButton).toBeVisible();
    await completeButton.click();
  }

  /**
   * Assign intervention to user
   * @param interventionTitle - Title of the intervention
   * @param assigneeName - Name of the assignee
   */
  async assignIntervention(interventionTitle: string, assigneeName: string): Promise<void> {
    const row = this.getRowByTitle(interventionTitle);
    const assignButton = row.locator(this.assignInterventionButton);
    await assignButton.click();
    
    // Select assignee from dropdown
    const assigneeOption = this.page.locator('[role="listbox"]', { hasText: new RegExp(assigneeName, 'i') });
    await assigneeOption.click();
  }

  /**
   * Get intervention status
   * @param interventionTitle - Title of the intervention
   */
  async getInterventionStatus(interventionTitle: string): Promise<string> {
    const row = this.getRowByTitle(interventionTitle);
    const statusCell = row.locator('[class*="status"], [class*="badge"]').first();
    await expect(statusCell).toBeVisible();
    return await statusCell.textContent() || '';
  }

  /**
   * Get intervention priority
   * @param interventionTitle - Title of the intervention
   */
  async getInterventionPriority(interventionTitle: string): Promise<string> {
    const row = this.getRowByTitle(interventionTitle);
    const priorityCell = row.locator('[class*="priority"]').first();
    await expect(priorityCell).toBeVisible();
    return await priorityCell.textContent() || '';
  }

  /**
   * Check if intervention exists
   * @param interventionTitle - Title to search for
   */
  async interventionExists(interventionTitle: string): Promise<boolean> {
    const row = this.getRowByTitle(interventionTitle);
    return await row.isVisible();
  }

  /**
   * Switch to Kanban board view
   */
  async switchToKanbanView(): Promise<void> {
    const kanbanButton = this.page.locator('button:has-text("Kanban"), [data-testid="kanban-view"]');
    await kanbanButton.click();
    await expect(this.kanbanBoard).toBeVisible({ timeout: 5000 });
  }

  /**
   * Switch to table view
   */
  async switchToTableView(): Promise<void> {
    const tableButton = this.page.locator('button:has-text("Table"), [data-testid="table-view"]');
    await tableButton.click();
    await expect(this.dataTable).toBeVisible({ timeout: 5000 });
  }

  /**
   * Get count of interventions in Kanban column
   * @param columnName - Name of the column (pending, in_progress, completed, blocked)
   */
  async getKanbanColumnCount(columnName: string): Promise<number> {
    let column: Locator;
    switch (columnName.toLowerCase()) {
      case 'pending':
        column = this.pendingColumn;
        break;
      case 'in_progress':
      case 'inprogress':
        column = this.inProgressColumn;
        break;
      case 'completed':
        column = this.completedColumn;
        break;
      case 'blocked':
        column = this.blockedColumn;
        break;
      default:
        throw new Error(`Unknown column: ${columnName}`);
    }
    
    const cards = column.locator('[class*="card"], [class*="intervention-item"]');
    return await cards.count();
  }

  /**
   * Add activity note to intervention
   * @param interventionTitle - Title of the intervention
   * @param note - Note text to add
   */
  async addNote(interventionTitle: string, note: string): Promise<void> {
    await this.viewIntervention(interventionTitle);
    await this.addNoteButton.click();
    
    const noteInput = this.page.locator('textarea, [contenteditable="true"]');
    await noteInput.fill(note);
    
    const submitButton = this.page.locator('button:has-text("Save"), button:has-text("Submit"), [type="submit"]');
    await submitButton.click();
  }

  /**
   * Update intervention status
   * @param interventionTitle - Title of the intervention
   * @param newStatus - New status to set
   */
  async updateStatus(interventionTitle: string, newStatus: string): Promise<void> {
    await this.viewIntervention(interventionTitle);
    await this.updateStatusButton.click();
    
    const statusOption = this.page.locator('[role="option"], option', { hasText: new RegExp(newStatus, 'i') });
    await statusOption.click();
  }
}
