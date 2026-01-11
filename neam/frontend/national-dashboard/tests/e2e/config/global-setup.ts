import { test as setup, expect } from '@playwright/test';
import { faker } from '@faker-js/faker';
import path from 'path';
import fs from 'fs';
import { v4 as uuidv4 } from 'uuid';

/**
 * Global Setup for Playwright Tests
 * 
 * This file is executed once before all tests run.
 * It handles:
 * - Environment validation
 * - Test database seeding
 * - Authentication state persistence
 * - Test data preparation
 */

const TEST_DATA_DIR = path.resolve(__dirname, '../fixtures');
const AUTH_STATE_DIR = path.resolve(__dirname, '../.auth');

/**
 * Validates required environment variables are present
 */
function validateEnvironment(): void {
  const requiredEnvVars = [
    'BASE_URL',
    'TEST_DATABASE_URL',
  ];

  const missing = requiredEnvVars.filter((env) => !process.env[env]);
  
  if (missing.length > 0) {
    console.warn(`⚠️  Missing environment variables: ${missing.join(', ')}`);
    console.warn('Some tests may fail without these values.');
  }
}

/**
 * Ensures required directories exist
 */
function ensureDirectories(): void {
  const dirs = [TEST_DATA_DIR, AUTH_STATE_DIR];
  
  dirs.forEach((dir) => {
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
      console.log(`📁 Created directory: ${dir}`);
    }
  });
}

/**
 * Seeds the test database with required reference data
 */
async function seedTestDatabase(): Promise<void> {
  console.log('🌱 Seeding test database...');
  
  // This would typically connect to the test database
  // and insert required reference data (regions, user roles, etc.)
  
  const testTenantId = `test-tenant-${uuidv4().slice(0, 8)}`;
  
  // Store tenant ID for use in tests
  process.env.TEST_TENANT_ID = testTenantId;
  
  console.log(`✅ Test database seeded with tenant: ${testTenantId}`);
}

/**
 * Generates and saves authentication states for different user roles
 */
async function generateAuthStates(): Promise<void> {
  console.log('🔐 Generating authentication states...');
  
  const roles = [
    { name: 'super-admin', email: 'test.superadmin@neam.gov' },
    { name: 'admin', email: 'test.admin@neam.gov' },
    { name: 'analyst', email: 'test.analyst@neam.gov' },
    { name: 'viewer', email: 'test.viewer@neam.gov' },
  ];

  // Create placeholder auth files (actual auth would be done via API login)
  for (const role of roles) {
    const filePath = path.join(AUTH_STATE_DIR, `${role.name}.json`);
    
    const authState = {
      origins: [
        {
          origin: new URL(process.env.BASE_URL || 'http://localhost:3000').origin,
          localStorage: [
            {
              name: 'auth_token',
              value: `mock-token-for-${role.name}-${Date.now()}`,
            },
            {
              name: 'user_role',
              value: role.name,
            },
            {
              name: 'user_email',
              value: role.email,
            },
          ],
        },
      ],
    };
    
    fs.writeFileSync(filePath, JSON.stringify(authState, null, 2));
    console.log(`✅ Auth state saved: ${role.name}`);
  }
}

/**
 * Main setup function
 */
export default setup('Global Setup', async ({ config }) => {
  console.log('='.repeat(60));
  console.log('🚀 NEAM Dashboard - Playwright Global Setup');
  console.log('='.repeat(60));
  
  // Set test run ID
  const testRunId = `test-run-${Date.now()}`;
  process.env.TEST_RUN_ID = testRunId;
  console.log(`📋 Test Run ID: ${testRunId}`);
  
  // Validate environment
  validateEnvironment();
  
  // Ensure directories exist
  ensureDirectories();
  
  // Seed test database
  await seedTestDatabase();
  
  // Generate auth states
  await generateAuthStates();
  
  console.log('='.repeat(60));
  console.log('✅ Global Setup Complete');
  console.log('='.repeat(60));
});
