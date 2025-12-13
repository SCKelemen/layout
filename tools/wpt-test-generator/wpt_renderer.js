#!/usr/bin/env node
/**
 * WPT Test Renderer - Universal JSON Schema Generator
 *
 * Renders WPT HTML tests in headless Chrome and outputs universal JSON test format.
 * Schema v1.0.0 - Usable by any implementation (Go, Rust, C, Python, etc.)
 *
 * Features:
 * - Multi-browser result support
 * - Declarative layout tree structure
 * - Proper categorization (categories, tags, properties)
 * - Source and spec metadata
 * - Only relevant properties per test type
 */

const puppeteer = require('puppeteer');
const fs = require('fs').promises;
const path = require('path');

/**
 * Parse CSS unit values to numbers (e.g., "100px" -> 100)
 */
function parsePixels(value) {
  if (!value || value === 'none' || value === 'auto') return null;
  const num = parseFloat(value);
  return isNaN(num) ? null : num;
}

/**
 * Detect categories based on computed styles and test structure
 */
function detectCategories(elements) {
  const categories = new Set();

  elements.forEach(el => {
    const display = el.computed.display;
    if (display === 'flex' || display === 'inline-flex') {
      categories.add('flexbox');
    }
    if (display === 'grid' || display === 'inline-grid') {
      categories.add('grid');
    }
  });

  // Always include 'layout' for now
  categories.add('layout');

  return Array.from(categories);
}

/**
 * Extract CSS properties being tested
 */
function detectProperties(elements) {
  const properties = new Set();

  elements.forEach(el => {
    const computed = el.computed;

    // Flex properties
    if (computed.display === 'flex' || computed.display === 'inline-flex') {
      if (computed.flexDirection !== 'row') properties.add('flex-direction');
      if (computed.flexWrap !== 'nowrap') properties.add('flex-wrap');
      if (computed.justifyContent !== 'normal' && computed.justifyContent !== 'flex-start') {
        properties.add('justify-content');
      }
      if (computed.alignItems !== 'normal' && computed.alignItems !== 'stretch') {
        properties.add('align-items');
      }
      if (computed.alignContent !== 'normal') {
        properties.add('align-content');
      }
    }

    // Grid properties
    if (computed.display === 'grid' || computed.display === 'inline-grid') {
      properties.add('display');
    }
  });

  // Default to basic sizing properties
  if (properties.size === 0) {
    properties.add('width');
    properties.add('height');
  }

  return Array.from(properties);
}

/**
 * Build declarative layout tree from extracted elements
 */
function buildLayoutTree(rootElement) {
  const style = {};
  const computed = rootElement.computed;

  // Display & Layout Mode
  if (computed.display) {
    const displayMap = {
      'flex': 'flex',
      'inline-flex': 'flex',
      'grid': 'grid',
      'inline-grid': 'grid',
      'block': 'block',
      'inline-block': 'inline-block'
    };
    style.display = displayMap[computed.display] || 'block';
  }

  // Flexbox properties
  if (computed.display === 'flex' || computed.display === 'inline-flex') {
    if (computed.flexDirection && computed.flexDirection !== 'row') {
      style.flexDirection = computed.flexDirection;
    }
    if (computed.flexWrap && computed.flexWrap !== 'nowrap') {
      style.flexWrap = computed.flexWrap;
    }
    if (computed.justifyContent && computed.justifyContent !== 'normal') {
      style.justifyContent = computed.justifyContent;
    }
    if (computed.alignItems && computed.alignItems !== 'normal') {
      style.alignItems = computed.alignItems;
    }
    if (computed.alignContent && computed.alignContent !== 'normal') {
      style.alignContent = computed.alignContent;
    }
  }

  // Box Model
  const width = parsePixels(computed.width);
  const height = parsePixels(computed.height);
  if (width) style.width = width;
  if (height) style.height = height;

  // Spacing
  const marginTop = parsePixels(computed.margin.top);
  const marginRight = parsePixels(computed.margin.right);
  const marginBottom = parsePixels(computed.margin.bottom);
  const marginLeft = parsePixels(computed.margin.left);

  if (marginTop || marginRight || marginBottom || marginLeft) {
    style.margin = {
      top: marginTop || 0,
      right: marginRight || 0,
      bottom: marginBottom || 0,
      left: marginLeft || 0
    };
  }

  const paddingTop = parsePixels(computed.padding.top);
  const paddingRight = parsePixels(computed.padding.right);
  const paddingBottom = parsePixels(computed.padding.bottom);
  const paddingLeft = parsePixels(computed.padding.left);

  if (paddingTop || paddingRight || paddingBottom || paddingLeft) {
    style.padding = {
      top: paddingTop || 0,
      right: paddingRight || 0,
      bottom: paddingBottom || 0,
      left: paddingLeft || 0
    };
  }

  // Build node
  const node = {
    type: 'container',
    id: rootElement.selector.replace('#', ''),
    style
  };

  // Add children
  if (rootElement.children && rootElement.children.length > 0) {
    node.children = rootElement.children.map((child, index) => {
      const childStyle = {};
      const childWidth = child.rect.width;
      const childHeight = child.rect.height;

      if (childWidth) childStyle.width = childWidth;
      if (childHeight) childStyle.height = childHeight;

      return {
        type: 'block',
        id: child.selector.replace('#', ''),
        style: childStyle
      };
    });
  }

  return node;
}

/**
 * Transform old format to new schema v1.0.0
 */
function transformToSchema(layoutData, inputFile) {
  const basename = path.basename(inputFile, path.extname(inputFile));
  const testTitle = layoutData.testFile || basename;

  // Find the root container (first flex/grid or first element)
  const rootElement = layoutData.elements.find(el =>
    el.computed.display === 'flex' ||
    el.computed.display === 'grid' ||
    el.computed.display === 'inline-flex' ||
    el.computed.display === 'inline-grid'
  ) || layoutData.elements[0];

  if (!rootElement) {
    throw new Error('No valid root element found');
  }

  const categories = detectCategories(layoutData.elements);
  const properties = detectProperties(layoutData.elements);
  const layoutTree = buildLayoutTree(rootElement);

  // Generate test ID from filename
  const testId = basename.replace(/[^a-z0-9-]/gi, '-').toLowerCase();

  // Build browser result
  const browserResult = {
    browser: {
      name: 'Chrome',
      version: layoutData.metadata?.browserVersion || 'Puppeteer Latest',
      engine: 'Blink'
    },
    rendered: {
      timestamp: layoutData.metadata?.generatedAt || new Date().toISOString(),
      viewport: layoutData.viewport
    },
    elements: [],
    tolerance: {
      position: 1.0,
      size: 1.0,
      numeric: 0.01
    }
  };

  // Add root element result
  browserResult.elements.push({
    id: rootElement.selector.replace('#', ''),
    path: 'root',
    expected: {
      x: rootElement.rect.x,
      y: rootElement.rect.y,
      width: rootElement.rect.width,
      height: rootElement.rect.height
    }
  });

  // Add children results
  if (rootElement.children) {
    rootElement.children.forEach((child, index) => {
      browserResult.elements.push({
        id: child.selector.replace('#', ''),
        path: `root.children[${index}]`,
        expected: {
          x: child.rect.x,
          y: child.rect.y,
          width: child.rect.width,
          height: child.rect.height
        }
      });
    });
  }

  // Build final schema
  return {
    version: '1.0.0',
    id: testId,
    title: testTitle,
    description: `Browser test for ${testTitle}`,
    source: {
      url: `file://${path.resolve(inputFile)}`,
      file: path.basename(inputFile),
      commit: null
    },
    generated: {
      timestamp: new Date().toISOString(),
      tool: 'wpt-test-generator v1.0.0'
    },
    spec: {
      name: categories.includes('flexbox') ? 'CSS Flexbox Level 1' : 'CSS',
      section: 'Layout',
      url: categories.includes('flexbox')
        ? 'https://www.w3.org/TR/css-flexbox-1/'
        : 'https://www.w3.org/TR/CSS/'
    },
    categories,
    tags: [],
    properties,
    layout: layoutTree,
    constraints: {
      type: 'loose',
      width: layoutData.viewport.width,
      height: layoutData.viewport.height
    },
    results: {
      chrome: browserResult
    },
    notes: []
  };
}

async function renderWPTTest(htmlFile, browserName = 'chrome') {
  const launchOptions = {
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  };

  // Add browser-specific configuration
  if (browserName === 'firefox') {
    launchOptions.product = 'firefox';
  }

  const browser = await puppeteer.launch(launchOptions);

  try {
    const page = await browser.newPage();

    // Set viewport size (standard test size)
    await page.setViewport({ width: 800, height: 600 });

    // Load the HTML file
    const htmlContent = await fs.readFile(htmlFile, 'utf-8');
    await page.setContent(htmlContent, { waitUntil: 'networkidle0' });

    // Extract layout information
    const layoutData = await page.evaluate(() => {
      const results = {
        testFile: document.title || 'Untitled Test',
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight
        },
        elements: []
      };

      // Find all elements - we'll filter by computed display later
      // Start with elements that have IDs or are divs (common test containers)
      const candidates = document.querySelectorAll('div, [id], section, article, main');

      const elementsToProcess = [];
      candidates.forEach((el) => {
        const computed = window.getComputedStyle(el);
        const display = computed.display;

        // Include flex/grid containers, or elements with IDs, or elements with data-expected-* attributes
        if (display === 'flex' || display === 'grid' ||
            display === 'inline-flex' || display === 'inline-grid' ||
            el.id ||
            el.hasAttribute('data-expected-width') ||
            el.hasAttribute('data-expected-height')) {
          elementsToProcess.push(el);
        }
      });

      elementsToProcess.forEach((el, index) => {
        const rect = el.getBoundingClientRect();
        const computed = window.getComputedStyle(el);

        // Skip if element has no size (invisible/collapsed)
        if (rect.width === 0 && rect.height === 0) {
          return;
        }

        const elementData = {
          selector: el.id ? `#${el.id}` : `element-${index}`,
          tagName: el.tagName.toLowerCase(),
          dataExpected: {},
          rect: {
            x: rect.x,
            y: rect.y,
            width: rect.width,
            height: rect.height,
            top: rect.top,
            left: rect.left,
            bottom: rect.bottom,
            right: rect.right
          },
          computed: {
            display: computed.display,
            position: computed.position,
            flexDirection: computed.flexDirection,
            flexWrap: computed.flexWrap,
            justifyContent: computed.justifyContent,
            alignItems: computed.alignItems,
            alignContent: computed.alignContent,
            width: computed.width,
            height: computed.height,
            minWidth: computed.minWidth,
            minHeight: computed.minHeight,
            maxWidth: computed.maxWidth,
            maxHeight: computed.maxHeight,
            margin: {
              top: computed.marginTop,
              right: computed.marginRight,
              bottom: computed.marginBottom,
              left: computed.marginLeft
            },
            padding: {
              top: computed.paddingTop,
              right: computed.paddingRight,
              bottom: computed.paddingBottom,
              left: computed.paddingLeft
            },
            border: {
              top: computed.borderTopWidth,
              right: computed.borderRightWidth,
              bottom: computed.borderBottomWidth,
              left: computed.borderLeftWidth
            }
          }
        };

        // Capture data-expected-* attributes if present
        ['width', 'height', 'client-width', 'client-height', 'offset-width', 'offset-height'].forEach(attr => {
          const dataAttr = `data-expected-${attr}`;
          if (el.hasAttribute(dataAttr)) {
            elementData.dataExpected[attr] = parseFloat(el.getAttribute(dataAttr));
          }
        });

        // Add children info if it's a flex/grid container
        if (computed.display === 'flex' || computed.display === 'grid' ||
            computed.display === 'inline-flex' || computed.display === 'inline-grid') {
          elementData.children = [];
          Array.from(el.children).forEach((child, childIndex) => {
            const childRect = child.getBoundingClientRect();
            if (childRect.width > 0 || childRect.height > 0) {
              elementData.children.push({
                selector: child.id ? `#${child.id}` : `child-${childIndex}`,
                rect: {
                  x: childRect.x,
                  y: childRect.y,
                  width: childRect.width,
                  height: childRect.height
                }
              });
            }
          });
        }

        results.elements.push(elementData);
      });

      return results;
    });

    return layoutData;
  } finally {
    await browser.close();
  }
}

async function processWPTFile(inputFile, outputFile, options = {}) {
  const browsers = options.browsers || ['chrome'];
  console.log(`Rendering ${inputFile} in ${browsers.join(' and ')}...`);

  try {
    // Render in first browser (typically Chrome)
    const chromeData = await renderWPTTest(inputFile, browsers[0]);

    // Add metadata
    chromeData.metadata = {
      generatedAt: new Date().toISOString(),
      sourceFile: path.basename(inputFile),
      browser: browsers[0] === 'firefox' ? 'Firefox Headless' : 'Chrome Headless',
      browserVersion: 'Puppeteer Latest'
    };

    // Transform to schema v1.0.0
    const schemaData = transformToSchema(chromeData, inputFile);

    // Render in additional browsers if specified
    for (let i = 1; i < browsers.length; i++) {
      const browserName = browsers[i];
      console.log(`  Rendering in ${browserName}...`);

      try {
        const browserData = await renderWPTTest(inputFile, browserName);
        browserData.metadata = {
          generatedAt: new Date().toISOString(),
          sourceFile: path.basename(inputFile),
          browser: browserName === 'firefox' ? 'Firefox Headless' : 'Chrome Headless',
          browserVersion: 'Puppeteer Latest'
        };

        // Add results for this browser
        const browserResult = buildBrowserResult(browserData, inputFile);
        schemaData.results[browserName] = browserResult;

        console.log(`  ✓ ${browserName} rendered (${browserResult.elements.length} elements)`);
      } catch (error) {
        console.warn(`  ⚠ Failed to render in ${browserName}: ${error.message}`);
        // Continue with other browsers
      }
    }

    // Write JSON output
    await fs.writeFile(
      outputFile,
      JSON.stringify(schemaData, null, 2),
      'utf-8'
    );

    const browserList = Object.keys(schemaData.results);
    console.log(`✓ Schema v1.0.0 generated`);
    console.log(`✓ Browsers: ${browserList.join(', ')}`);
    console.log(`✓ Categories: ${schemaData.categories.join(', ')}`);
    console.log(`✓ Properties: ${schemaData.properties.join(', ')}`);
    console.log(`✓ Saved to ${outputFile}`);

    return schemaData;
  } catch (error) {
    console.error(`Error processing ${inputFile}:`, error.message);
    throw error;
  }
}

/**
 * Build a browser result object from layout data
 */
function buildBrowserResult(layoutData, inputFile) {
  // Find the root container
  const rootElement = layoutData.elements.find(el =>
    el.computed.display === 'flex' ||
    el.computed.display === 'grid' ||
    el.computed.display === 'inline-flex' ||
    el.computed.display === 'inline-grid'
  ) || layoutData.elements[0];

  if (!rootElement) {
    throw new Error('No valid root element found');
  }

  const browserResult = {
    browser: {
      name: layoutData.metadata.browser.includes('Firefox') ? 'Firefox' : 'Chrome',
      version: layoutData.metadata.browserVersion || 'Puppeteer Latest',
      engine: layoutData.metadata.browser.includes('Firefox') ? 'Gecko' : 'Blink'
    },
    rendered: {
      timestamp: layoutData.metadata?.generatedAt || new Date().toISOString(),
      viewport: layoutData.viewport
    },
    elements: [],
    tolerance: {
      position: 1.0,
      size: 1.0,
      numeric: 0.01
    }
  };

  // Add root element result
  browserResult.elements.push({
    id: rootElement.selector.replace('#', ''),
    path: 'root',
    expected: {
      x: rootElement.rect.x,
      y: rootElement.rect.y,
      width: rootElement.rect.width,
      height: rootElement.rect.height
    }
  });

  // Add children results
  if (rootElement.children) {
    rootElement.children.forEach((child, index) => {
      browserResult.elements.push({
        id: child.selector.replace('#', ''),
        path: `root.children[${index}]`,
        expected: {
          x: child.rect.x,
          y: child.rect.y,
          width: child.rect.width,
          height: child.rect.height
        }
      });
    });
  }

  return browserResult;
}

async function batchProcess(inputDir, outputDir, options = {}) {
  const files = await fs.readdir(inputDir);
  const htmlFiles = files.filter(f => f.endsWith('.html') || f.endsWith('.htm'));

  const browsers = options.browsers || ['chrome'];
  console.log(`Found ${htmlFiles.length} HTML test files`);
  console.log(`Using browsers: ${browsers.join(', ')}`);

  await fs.mkdir(outputDir, { recursive: true });

  const results = [];
  for (const file of htmlFiles) {
    const inputPath = path.join(inputDir, file);
    const outputPath = path.join(outputDir, file.replace(/\.html?$/, '.json'));

    try {
      const data = await processWPTFile(inputPath, outputPath, options);
      const firstBrowser = Object.keys(data.results)[0];
      results.push({
        file,
        status: 'success',
        testId: data.id,
        browsers: Object.keys(data.results),
        categories: data.categories,
        properties: data.properties,
        elements: data.results[firstBrowser].elements.length
      });
    } catch (error) {
      results.push({ file, status: 'error', error: error.message });
    }
  }

  // Write summary
  const summary = {
    processedAt: new Date().toISOString(),
    totalFiles: htmlFiles.length,
    successful: results.filter(r => r.status === 'success').length,
    failed: results.filter(r => r.status === 'error').length,
    results
  };

  await fs.writeFile(
    path.join(outputDir, '_summary.json'),
    JSON.stringify(summary, null, 2),
    'utf-8'
  );

  console.log('\n=== Summary ===');
  console.log(`Total: ${summary.totalFiles}`);
  console.log(`Success: ${summary.successful}`);
  console.log(`Failed: ${summary.failed}`);

  return summary;
}

// CLI
async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    console.error('Usage:');
    console.error('  Single file: node wpt_renderer.js <input.html> [output.json] [--browsers chrome,firefox]');
    console.error('  Batch:       node wpt_renderer.js --batch <input-dir> <output-dir> [--browsers chrome,firefox]');
    console.error('');
    console.error('Options:');
    console.error('  --browsers   Comma-separated list of browsers (chrome, firefox)');
    console.error('               Default: chrome');
    process.exit(1);
  }

  // Parse browsers option
  let browsers = ['chrome'];
  const browsersIndex = args.indexOf('--browsers');
  if (browsersIndex !== -1 && args[browsersIndex + 1]) {
    browsers = args[browsersIndex + 1].split(',').map(b => b.trim());
  }

  if (args[0] === '--batch') {
    if (args.length < 3) {
      console.error('Batch mode requires input and output directories');
      process.exit(1);
    }
    await batchProcess(args[1], args[2], { browsers });
  } else {
    const inputFile = args[0];
    const outputFile = args[1] && !args[1].startsWith('--')
      ? args[1]
      : inputFile.replace(/\.html?$/, '.json');
    await processWPTFile(inputFile, outputFile, { browsers });
  }
}

if (require.main === module) {
  main().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });
}

module.exports = { renderWPTTest, processWPTFile, batchProcess };
