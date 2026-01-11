import { defineConfig, devices, PlaywrightTestConfig } from '@playwright/test';
import path from 'path';
import dotenv from 'dotenv';

// Load environment variables from .env.test.local or .env.test
dotenv.config({ path: path.resolve(__dirname, '.env.test') });

const PORT = process.env.PORT || 3000;
const BASE_URL = process.env.BASE_URL || `http://localhost:${PORT}`;

/**
 * Playwright Test Configuration for NEAM National Dashboard
 * 
 * This configuration sets up the testing environment for:
 * - E2E workflow testing
 * - Multi-role approval workflows
 * - Performance baseline testing
 */
export default defineConfig({
  // Directory where tests are located
  testDir: './tests/e2e',
  
  // Files to ignore
  testIgnore: [
    '**/*.css',
    '**/*.md',
    'node_modules',
    'dist',
    'build'
  ],
  
  // Global timeout for all tests (30 minutes for long E2E flows)
  timeout: 30 * 1000,
  
  // Expect timeout for assertions
  expect: {
    timeout: 10000,
    // Custom message for soft assertions
    toHaveScreenshot: {
      maxDiffPixels: 100,
    },
  },
  
  // Fully parallelize tests
  fullyParallel: true,
  
  // Fail tests on flakiness
  forbidOnly: !!process.env.CI,
  
  // Retry failed tests
  retries: process.env.CI ? 2 : 0,
  
  // Workers for parallel execution
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter configuration
  reporter: [
    ['html', { 
      outputFolder: './test-results/html',
      open: 'never'
    }],
    ['json', { 
      outputFile: './test-results/results.json'
    }],
    ['list', { 
      printSteps: true 
    }],
    ['allure', {
      outputFolder: './test-results/allure',
    }]
  ],
  
  // Shared settings for all projects
  use: {
    // Base URL for all tests
    baseURL: BASE_URL,
    
    // Trace recording options
    trace: 'on-first-retry',
    
    // Screenshot options
    screenshot: 'only-on-failure',
    
    // Video recording
    video: 'retain-on-failure',
    
    // Locale
    locale: 'en-US',
    
    // Timezone
    timezoneId: 'UTC',
    
    // User agent
    userAgent: 'NEAM-Dashboard-Test/1.0',
    
    // Custom headers for API requests
    extraHTTPHeaders: {
      'X-Test-Environment': 'e2e',
      'X-Request-ID': process.env.TEST_RUN_ID || 'test-run',
    },
  },
  
  // Configure projects for different browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    // Mobile testing
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'mobile-safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  
  // Web server configuration
  webServer: {
    command: process.env.SKIP_BUILD 
      ? `npm run dev` 
      : `npm run build && npm run start`,
    url: BASE_URL,
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
    // Environment variables for web server
    env: {
      NODE_ENV: 'test',
      TEST_DATABASE_URL: process.env.TEST_DATABASE_URL,
      TEST_REDIS_URL: process.env.TEST_REDIS_URL,
    },
  },
  
  // Global setup and teardown
  globalSetup: './config/global-setup.ts',
  globalTeardown: './config/global-teardown.ts',
  
  // Dependency projects (run these first)
  dependencies: [],
  
  // Shard tests across multiple machines
  shards: undefined,
  
  // Maximum memory for the test runner
  use: {
    ...devices['Desktop Chrome'],
  },
} as PlaywrightTestConfig);

// Export base URL for use in other files
export const getBaseUrl = () => BASE_URL;
export const getTestRunId = () => process.env.TEST_RUN_ID || `test-run-${Date.now()}`;
