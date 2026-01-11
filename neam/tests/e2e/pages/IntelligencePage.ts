import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Intelligence page
 * Handles all intelligence item management interactions and verifications
 */
export class IntelligencePage {
  readonly page: Page;
  
  // Page header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;
  
  // Action buttons
  readonly createIntelligenceButton: Locator;
  readonly importButton: Locator;
  readonly exportButton: Locator;
  readonly refreshButton: Locator;
  
  // Search and filters
  readonly searchInput: Locator;
  readonly statusFilter: Locator;
  readonly severityFilter: Locator;
  readonly sourceFilter: Locator;
  readonly dateRangeFilter: Locator;
  readonly clearFiltersButton: Locator;
  
  // Data table
  readonly dataTable: Locator;
  readonly tableHeaders: Locator;
  readonly tableRows: Locator;
  readonly tableCells: Locator;
  
  // Pagination
  readonly paginationInfo: Locator;
  readonly previousPageButton: Locator;
  readonly nextPageButton: Locator;
  readonly pageNumbers: Locator;
  
  // Individual intelligence item actions
  readonly viewItemButton: Locator;
  readonly editItemButton: Locator;
  readonly deleteItemButton: Locator;
  readonly promoteToInterventionButton: Locator;
  
  // Status badges
  readonly newStatusBadge: Locator;
  readonly pendingStatusBadge: Locator;
  readonly underReviewStatusBadge: Locator;
  readonly resolvedStatusBadge: Locator;
  
  // Severity badges
  readonly lowSeverityBadge: Locator;
  readonly mediumSeverityBadge: Locator;
  readonly highSeverityBadge: Locator;
  readonly criticalSeverityBadge: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Page header
    this.pageTitle = page.locator('[data-testid="page-title"], h1:has-text("Intelligence"), .page-title');
    this.pageDescription = page.locator('[data-testid="page-description"], .page-description, p');
    
    // Actions
    this.createIntelligenceButton = page.locator('[data-testid="create-intelligence-button"], button:has-text("Create Intelligence"), button:has-text("New Intelligence"), [class*="create"][class*="intelligence"]');
    this.importButton = page.locator('[data-testid="import-button"], button:has-text("Import"), [class*="import"]');
    this.exportButton = page.locator('[data-testid="export-button"], button:has-text("Export"), [class*="export"]');
    this.refreshButton = page.locator('[data-testid="refresh-button"], button:has-text("Refresh"), [class*="refresh"]');
    
    // Filters
    this.searchInput = page.locator('[data-testid="search-input"], input[type="search"], input[placeholder*="Search"]');
    this.statusFilter = page.locator('[data-testid="status-filter"], [class*="status-filter"], select:has-text("Status")');
    this.severityFilter = page.locator('[data-testid="severity-filter"], [class*="severity-filter"], select:has-text("Severity")');
    this.sourceFilter = page.locator('[data-testid="source-filter"], [class*="source-filter"], select:has-text("Source")');
    this.dateRangeFilter = page.locator('[data-testid="date-range-filter"], [class*="date-filter"], input[type="date"]');
    this.clearFiltersButton = page.locator('[data-testid="clear-filters-button"], button:has-text("Clear Filters"), [class*="clear-filter"]');
    
    // Data table
    this.dataTable = page.locator('[data-testid="intelligence-table"], table, [class*="data-table"]');
    this.tableHeaders = page.locator('[data-testid="intelligence-table"] th, table th, [class*="table-header"]');
    this.tableRows = page.locator('[data-testid="intelligence-table"] tbody tr, table tbody tr, [class*="table-row"]:not([class*="header"])');
    this.tableCells = page.locator('[data-testid="intelligence-table"] td, table td, [class*="table-cell"]');
    
    // Pagination
    this.paginationInfo = page.locator('[data-testid="pagination-info"], .pagination-info, [class*="pagination"] span');
    this.previousPageButton = page.locator('[data-testid="previous-page-button"], button:has-text("Previous"), [class*="previous-page"]');
    this.nextPageButton = page.locator('[data-testid="next-page-button"], button:has-text("Next"), [class*="next-page"]');
    this.pageNumbers = page.locator('[data-testid="page-number"], [class*="page-number"]:not([class*="button"])');
    
    // Item actions (generic - will be scoped to specific rows)
    this.viewItemButton = page.locator('[data-testid="view-button"], button:has-text("View"), a:has-text("View"), [class*="view"]');
    this.editItemButton = page.locator('[data-testid="edit-button"], button:has-text("Edit"), [class*="edit"]');
    this.deleteItemButton = page.locator('[data-testid="delete-button"], button:has-text("Delete"), [class*="delete"]');
    this.promoteToInterventionButton = page.locator('[data-testid="promote-button"], button:has-text("Promote"), button:has-text("Create Intervention"), [class*="promote"]');
    
    // Status badges
    this.newStatusBadge = page.locator('[data-testid="status-new"], [class*="status"]:has-text("New")');
    this.pendingStatusBadge = page.locator('[data-testid="status-pending"], [class*="status"]:has-text("Pending")');
    this.underReviewStatusBadge = page.locator('[data-testid="status-under-review"], [class*="status"]:has-text("Under Review")');
    this.resolvedStatusBadge = page.locator('[data-testid="status-resolved"], [class*="status"]:has-text("Resolved")');
    
    // Severity badges
    this.lowSeverityBadge = page.locator('[data-testid="severity-low"], [class*="severity"]:has-text("Low")');
    this.mediumSeverityBadge = page.locator('[data-testid="severity-medium"], [class*="severity"]:has-text("Medium")');
    this.highSeverityBadge = page.locator('[data-testid="severity-high"], [class*="severity"]:has-text("High")');
    this.criticalSeverityBadge = page.locator('[data-testid="severity-critical"], [class*="severity"]:has-text("Critical")');
  }

  /**
   * Navigate to intelligence page
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/intelligence`, { waitUntil: 'networkidle' });
  }

  /**
   * Wait for page to fully load
   */
  async waitForPageLoad(): Promise<void> {
    await expect(this.dataTable).toBeVisible({ timeout: 10000 });
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
   * Get total count of intelligence items
   */
  async getItemsCount(): Promise<number> {
    return await this.tableRows.count();
  }

  /**
   * Search for intelligence items
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
   * Filter by severity
   * @param severity - Severity level to filter by
   */
  async filterBySeverity(severity: string): Promise<void> {
    await this.severityFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(severity, 'i') });
    await option.click();
    await this.waitForPageLoad();
  }

  /**
   * Filter by source
   * @param source - Source to filter by
   */
  async filterBySource(source: string): Promise<void> {
    await this.sourceFilter.click();
    const option = this.page.locator('option', { hasText: new RegExp(source, 'i') });
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
   * Click create new intelligence button
   */
  async clickCreateIntelligence(): Promise<void> {
    await this.createIntelligenceButton.click();
  }

  /**
   * Click import button
   */
  async clickImport(): Promise<void> {
    await this.importButton.click();
  }

  /**
   * Click export button
   */
  async clickExport(): Promise<void> {
    await this.exportButton.click();
  }

  /**
   * Click refresh button
   */
  async clickRefresh(): Promise<void> {
    await this.refreshButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Click view button for specific item
   * @param itemTitle - Title of the item to view
   */
  async viewItem(itemTitle: string): Promise<void> {
    const row = this.getRowByTitle(itemTitle);
    const viewButton = row.locator(this.viewItemButton);
    await expect(viewButton).toBeVisible();
    await viewButton.click();
  }

  /**
   * Click edit button for specific item
   * @param itemTitle - Title of the item to edit
   */
  async editItem(itemTitle: string): Promise<void> {
    const row = this.getRowByTitle(itemTitle);
    const editButton = row.locator(this.editItemButton);
    await expect(editButton).toBeVisible();
    await editButton.click();
  }

  /**
   * Click delete button for specific item
   * @param itemTitle - Title of the item to delete
   */
  async deleteItem(itemTitle: string): Promise<void> {
    const row = this.getRowByTitle(itemTitle);
    const deleteButton = row.locator(this.deleteItemButton);
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();
    
    // Confirm deletion in dialog
    const confirmButton = this.page.locator('button:has-text("Delete"), button:has-text("Confirm"), [data-testid="confirm-delete"]');
    await confirmButton.click();
  }

  /**
   * Promote item to intervention
   * @param itemTitle - Title of the item to promote
   */
  async promoteToIntervention(itemTitle: string): Promise<void> {
    const row = this.getRowByTitle(itemTitle);
    const promoteButton = row.locator(this.promoteToInterventionButton);
    await expect(promoteButton).toBeVisible();
    await promoteButton.click();
  }

  /**
   * Get row by item title
   * @param itemTitle - Title to search for
   */
  private getRowByTitle(itemTitle: string): Locator {
    return this.tableRows.filter({ has: this.page.locator('td', { hasText: new RegExp(itemTitle, 'i') }) });
  }

  /**
   * Get item status
   * @param itemTitle - Title of the item
   */
  async getItemStatus(itemTitle: string): Promise<string> {
    const row = this.getRowByTitle(itemTitle);
    const statusCell = row.locator('[class*="status"], [class*="badge"]').first();
    await expect(statusCell).toBeVisible();
    return await statusCell.textContent() || '';
  }

  /**
   * Get item severity
   * @param itemTitle - Title of the item
   */
  async getItemSeverity(itemTitle: string): Promise<string> {
    const row = this.getRowByTitle(itemTitle);
    const severityCell = row.locator('[class*="severity"]').first();
    await expect(severityCell).toBeVisible();
    return await severityCell.textContent() || '';
  }

  /**
   * Click next page button
   */
  async clickNextPage(): Promise<void> {
    await this.nextPageButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Click previous page button
   */
  async clickPreviousPage(): Promise<void> {
    await this.previousPageButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Go to specific page
   * @param pageNumber - Page number to navigate to
   */
  async goToPage(pageNumber: number): Promise<void> {
    const pageButton = this.pageNumbers.filter({ hasText: pageNumber.toString() });
    await pageButton.click();
    await this.waitForPageLoad();
  }

  /**
   * Get current page number
   */
  async getCurrentPage(): Promise<number> {
    const activePageButton = this.pageNumbers.locator('[class*="active"], [aria-current="page"]');
    const text = await activePageButton.textContent();
    return parseInt(text || '1', 10);
  }

  /**
   * Check if specific item exists in table
   * @param itemTitle - Title to search for
   */
  async itemExists(itemTitle: string): Promise<boolean> {
    const row = this.getRowByTitle(itemTitle);
    return await row.isVisible();
  }

  /**
   * Get table headers
   */
  async getTableHeaders(): Promise<string[]> {
    const headers = await this.tableHeaders.allTextContents();
    return headers.map(h => h.trim());
  }

  /**
   * Sort by column
   * @param columnName - Name of the column to sort by
   */
  async sortByColumn(columnName: string): Promise<void> {
    const header = this.tableHeaders.filter({ hasText: new RegExp(columnName, 'i') });
    await header.click();
    await this.waitForPageLoad();
  }
}
