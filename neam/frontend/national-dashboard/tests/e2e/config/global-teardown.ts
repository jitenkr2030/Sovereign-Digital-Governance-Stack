import { test as teardown } from '@playwright/test';

/**
 * Global Teardown for Playwright Tests
 * 
 * This file is executed once after all tests complete.
 * It handles:
 * - Cleanup of test data
 * - Database teardown
 * - Test artifacts cleanup
 */

async function cleanupTestData(): Promise<void> {
  console.log('🧹 Cleaning up test data...');
  
  // This would typically:
  // 1. Clean up test tenant data
  // 2. Remove test files
  // 3. Close database connections
  
  const testTenantId = process.env.TEST_TENANT_ID;
  
  if (testTenantId) {
    console.log(`🗑️  Removing test tenant data: ${testTenantId}`);
    // API call to clean up test data would go here
  }
}

async function cleanupTestArtifacts(): Promise<void> {
  console.log('📦 Cleaning up test artifacts...');
  
  // Keep HTML reports, remove video traces on success
  // This can be configured based on test results
}

async function generateTestSummary(): Promise<void> {
  console.log('📊 Generating test summary...');
  
  // Read test results and generate summary
  const testResultsPath = './test-results/results.json';
  
  if (process.env.fs?.existsSync(testResultsPath)) {
    try {
      const results = JSON.parse(process.env.fs?.readFileSync(testResultsPath, 'utf-8'));
      
      const stats = {
        total: results.stats?.tests || 0,
        passed: results.stats?.passed || 0,
        failed: results.stats?.failed || 0,
        skipped: results.stats?.skipped || 0,
        duration: results.stats?.duration || 0,
      };
      
      console.log('\n📈 Test Summary:');
      console.log(`   Total: ${stats.total}`);
      console.log(`   Passed: ${stats.passed}`);
      console.log(`   Failed: ${stats.failed}`);
      console.log(`   Skipped: ${stats.skipped}`);
      console.log(`   Duration: ${(stats.duration / 1000).toFixed(2)}s`);
    } catch (error) {
      console.log('⚠️  Could not read test results');
    }
  }
}

/**
 * Main teardown function
 */
export default teardown('Global Teardown', async ({ config }) => {
  console.log('\n' + '='.repeat(60));
  console.log('🏁 NEAM Dashboard - Global Teardown');
  console.log('='.repeat(60));
  
  try {
    // Cleanup test data
    await cleanupTestData();
    
    // Cleanup artifacts
    await cleanupTestArtifacts();
    
    // Generate summary
    await generateTestSummary();
    
    console.log('\n✅ Global Teardown Complete');
  } catch (error) {
    console.error('❌ Error during teardown:', error);
  }
});
