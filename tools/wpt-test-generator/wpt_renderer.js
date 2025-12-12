#!/usr/bin/env node
/**
 * WPT Test Renderer - Headless Chrome
 *
 * Renders WPT HTML tests in headless Chrome and extracts:
 * - Element positions (getBoundingClientRect)
 * - Computed styles (getComputedStyle)
 * - Layout dimensions
 *
 * Outputs JSON with expected values for Go test generation.
 */

const puppeteer = require('puppeteer');
const fs = require('fs').promises;
const path = require('path');

async function renderWPTTest(htmlFile) {
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

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

async function processWPTFile(inputFile, outputFile) {
  console.log(`Rendering ${inputFile} in headless Chrome...`);

  try {
    const layoutData = await renderWPTTest(inputFile);

    // Add metadata
    layoutData.metadata = {
      generatedAt: new Date().toISOString(),
      sourceFile: path.basename(inputFile),
      browser: 'Chrome Headless',
      browserVersion: 'Puppeteer Latest'
    };

    // Write JSON output
    await fs.writeFile(
      outputFile,
      JSON.stringify(layoutData, null, 2),
      'utf-8'
    );

    console.log(`✓ Layout data extracted: ${layoutData.elements.length} elements`);
    console.log(`✓ Saved to ${outputFile}`);

    return layoutData;
  } catch (error) {
    console.error(`Error processing ${inputFile}:`, error.message);
    throw error;
  }
}

async function batchProcess(inputDir, outputDir) {
  const files = await fs.readdir(inputDir);
  const htmlFiles = files.filter(f => f.endsWith('.html') || f.endsWith('.htm'));

  console.log(`Found ${htmlFiles.length} HTML test files`);

  await fs.mkdir(outputDir, { recursive: true });

  const results = [];
  for (const file of htmlFiles) {
    const inputPath = path.join(inputDir, file);
    const outputPath = path.join(outputDir, file.replace(/\.html?$/, '.json'));

    try {
      const data = await processWPTFile(inputPath, outputPath);
      results.push({ file, status: 'success', elements: data.elements.length });
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
    console.error('  Single file: node wpt_renderer.js <input.html> [output.json]');
    console.error('  Batch:       node wpt_renderer.js --batch <input-dir> <output-dir>');
    process.exit(1);
  }

  if (args[0] === '--batch') {
    if (args.length < 3) {
      console.error('Batch mode requires input and output directories');
      process.exit(1);
    }
    await batchProcess(args[1], args[2]);
  } else {
    const inputFile = args[0];
    const outputFile = args[1] || inputFile.replace(/\.html?$/, '.json');
    await processWPTFile(inputFile, outputFile);
  }
}

if (require.main === module) {
  main().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });
}

module.exports = { renderWPTTest, processWPTFile, batchProcess };
