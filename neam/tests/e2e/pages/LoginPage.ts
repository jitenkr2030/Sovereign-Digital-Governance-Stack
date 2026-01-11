import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page Object Model for the Login page
 * Handles all login-related interactions and verifications
 */
export class LoginPage {
  readonly page: Page;
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly forgotPasswordLink: Locator;
  readonly rememberMeCheckbox: Locator;
  readonly errorMessage: Locator;
  readonly logo: Locator;
  readonly loginForm: Locator;
  readonly loadingSpinner: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Form elements
    this.usernameInput = page.locator('[data-testid="username-input"], input[type="email"], input[name="email"], input[id*="email"]');
    this.passwordInput = page.locator('[data-testid="password-input"], input[type="password"], input[name="password"], input[id*="password"]');
    this.loginButton = page.locator('[data-testid="login-button"], button[type="submit"], button:has-text("Sign In"), button:has-text("Login")');
    this.forgotPasswordLink = page.locator('[data-testid="forgot-password-link"], a:has-text("Forgot Password"), a:has-text("Forgot your password")');
    this.rememberMeCheckbox = page.locator('[data-testid="remember-me"], input[type="checkbox"], label:has-text("Remember me")');
    this.errorMessage = page.locator('[data-testid="error-message"], .error-message, .alert-error, [role="alert"]');
    
    // Branding elements
    this.logo = page.locator('[data-testid="app-logo"], .logo, img[alt*="logo"], svg[class*="logo"]');
    this.loginForm = page.locator('[data-testid="login-form"], form, .login-form');
    
    // State indicators
    this.loadingSpinner = page.locator('[data-testid="loading-spinner"], .spinner, [class*="loading"]');
  }

  /**
   * Navigate to the login page
   * @param baseURL - Base URL of the application
   */
  async navigateTo(baseURL: string): Promise<void> {
    await this.page.goto(`${baseURL}/login`, { waitUntil: 'networkidle' });
  }

  /**
   * Enter username/email
   * @param username - Username or email to enter
   */
  async enterUsername(username: string): Promise<void> {
    await this.usernameInput.clear();
    await this.usernameInput.fill(username);
  }

  /**
   * Enter password
   * @param password - Password to enter
   */
  async enterPassword(password: string): Promise<void> {
    await this.passwordInput.clear();
    await this.passwordInput.fill(password);
  }

  /**
   * Click the login button
   */
  async clickLoginButton(): Promise<void> {
    await this.loginButton.click();
  }

  /**
   * Click forgot password link
   */
  async clickForgotPassword(): Promise<void> {
    await this.forgotPasswordLink.click();
  }

  /**
   * Toggle remember me checkbox
   */
  async toggleRememberMe(): Promise<void> {
    await this.rememberMeCheckbox.click();
  }

  /**
   * Perform complete login flow
   * @param username - Username or email
   * @param password - Password
   * @param rememberMe - Whether to check remember me (optional)
   */
  async login(username: string, password: string, rememberMe?: boolean): Promise<void> {
    await this.enterUsername(username);
    await this.enterPassword(password);
    
    if (rememberMe) {
      await this.toggleRememberMe();
    }
    
    await this.clickLoginButton();
  }

  /**
   * Wait for login to complete and redirect to dashboard
   */
  async waitForLoginSuccess(): Promise<void> {
    // Wait for loading spinner to disappear if present
    await this.page.waitForSelector(this.loadingSpinner, { state: 'hidden' }).catch(() => {});
    
    // Wait for navigation or URL change
    await this.page.waitForURL('**/dashboard**', { timeout: 10000 }).catch(() => {
      // Alternative: wait for specific dashboard element
      const dashboard = this.page.locator('[data-testid="dashboard"], .dashboard, nav:has-text("Dashboard")');
      expect(dashboard.first()).toBeVisible({ timeout: 5000 });
    });
  }

  /**
   * Check if error message is displayed
   */
  async getErrorMessage(): Promise<string | null> {
    const errorElement = this.errorMessage.first();
    if (await errorElement.isVisible()) {
      return await errorElement.textContent();
    }
    return null;
  }

  /**
   * Verify login page is displayed correctly
   */
  async verifyPageLoaded(): Promise<void> {
    await expect(this.usernameInput).toBeVisible();
    await expect(this.passwordInput).toBeVisible();
    await expect(this.loginButton).toBeVisible();
  }

  /**
   * Check if login button is disabled
   */
  async isLoginButtonDisabled(): Promise<boolean> {
    const isDisabled = await this.loginButton.getAttribute('disabled');
    const isAriaDisabled = await this.loginButton.getAttribute('aria-disabled');
    
    return (isDisabled !== null && isDisabled !== '') || 
           (isAriaDisabled === 'true') ||
           !(await this.loginButton.isEnabled());
  }

  /**
   * Get page title
   */
  async getPageTitle(): Promise<string> {
    return await this.page.title();
  }

  /**
   * Check if logo is visible
   */
  async isLogoVisible(): Promise<boolean> {
    return await this.logo.first().isVisible();
  }

  /**
   * Wait for loading state to complete
   */
  async waitForLoadingToComplete(): Promise<void> {
    // Wait for spinner to appear then disappear
    try {
      await this.page.waitForSelector(this.loadingSpinner, { state: 'visible', timeout: 2000 });
      await this.page.waitForSelector(this.loadingSpinner, { state: 'hidden', timeout: 10000 });
    } catch {
      // Spinner might not appear, that's okay
    }
  }

  /**
   * Login as admin user (convenience method)
   */
  async loginAsAdmin(): Promise<void> {
    await this.login(process.env.TEST_ADMIN_EMAIL || 'admin@test.gov', process.env.TEST_ADMIN_PASSWORD || 'test_password123');
  }

  /**
   * Login as analyst user (convenience method)
   */
  async loginAsAnalyst(): Promise<void> {
    await this.login(process.env.TEST_ANALYST_EMAIL || 'analyst@test.gov', process.env.TEST_ANALYST_PASSWORD || 'test_password123');
  }

  /**
   * Login as supervisor user (convenience method)
   */
  async loginAsSupervisor(): Promise<void> {
    await this.login(process.env.TEST_SUPERVISOR_EMAIL || 'supervisor@test.gov', process.env.TEST_SUPERVISOR_PASSWORD || 'test_password123');
  }
}
