import { chromium } from 'playwright';

async function testCSICPlatform() {
  console.log('Starting CSIC Platform test...');

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 }
  });

  const page = await context.newPage();

  // Collect console errors
  const consoleErrors = [];
  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });

  page.on('pageerror', error => {
    consoleErrors.push(error.message);
  });

  try {
    // Navigate to the app (we'll start the dev server first)
    console.log('Navigating to http://localhost:3000...');
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: 30000 });

    console.log('Page loaded successfully');

    // Wait for the app to initialize
    await page.waitForTimeout(2000);

    // Check if we're on the login page (expected since user is not authenticated)
    const pageTitle = await page.title();
    console.log('Page title:', pageTitle);

    // Check for main UI elements
    const bodyContent = await page.textContent('body');
    console.log('Page contains content:', bodyContent.length > 0);

    // Check for specific elements
    const hasRoot = await page.$('#root');
    console.log('Root element present:', !!hasRoot);

    // Report console errors
    if (consoleErrors.length > 0) {
      console.log('\nConsole errors found:');
      consoleErrors.forEach((error, i) => {
        console.log(`${i + 1}. ${error}`);
      });
    } else {
      console.log('\nNo console errors found!');
    }

    console.log('\nTest completed successfully!');
    return consoleErrors.length === 0;

  } catch (error) {
    console.error('Test failed:', error.message);
    return false;
  } finally {
    await browser.close();
  }
}

// Run the test
testCSICPlatform()
  .then(success => {
    process.exit(success ? 0 : 1);
  })
  .catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });
