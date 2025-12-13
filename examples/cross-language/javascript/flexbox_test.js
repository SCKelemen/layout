#!/usr/bin/env node
/**
 * JavaScript example: Testing layout with wptest eval
 *
 * This demonstrates how to use wptest from JavaScript without any dependencies
 * beyond Node.js standard library. Just spawn wptest and pass JSON via stdin.
 */

const { spawn } = require('child_process');
const assert = require('assert');

// Test spec - same format works in ANY language
const testSpec = {
  layout: {
    type: 'container',
    style: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      width: 600,
      height: 100,
    },
    children: [
      { type: 'container', style: { width: 100, height: 50 } },
      { type: 'container', style: { width: 100, height: 50 } },
      { type: 'container', style: { width: 100, height: 50 } },
    ],
  },
  constraints: {
    maxWidth: 800,
    maxHeight: 600,
  },
  assertions: [
    {
      type: 'layout',
      expression: 'getX(child(root(), 0)) == 0.0',
      message: 'first-child-at-start',
    },
    {
      type: 'layout',
      expression: 'getRight(child(root(), 2)) == getWidth(root())',
      message: 'last-child-at-end',
    },
    {
      type: 'layout',
      expression: 'getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0',
      message: 'vertically-centered',
    },
  ],
  binding: 'old',
};

// Spawn wptest and communicate via stdin/stdout
function runLayoutTest(spec) {
  return new Promise((resolve, reject) => {
    const wptest = spawn('wptest', ['eval']);
    let stdout = '';
    let stderr = '';

    wptest.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    wptest.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    wptest.on('close', (code) => {
      if (code !== 0) {
        reject(new Error(`wptest exited with code ${code}\n${stderr}`));
      } else {
        try {
          resolve(JSON.parse(stdout));
        } catch (err) {
          reject(new Error(`Failed to parse JSON: ${err.message}\n${stdout}`));
        }
      }
    });

    wptest.on('error', (err) => {
      reject(new Error(`Failed to spawn wptest: ${err.message}`));
    });

    // Write test spec to stdin
    wptest.stdin.write(JSON.stringify(spec));
    wptest.stdin.end();
  });
}

// Run the test
async function main() {
  console.log('Running flexbox layout test from JavaScript...\n');

  try {
    const result = await runLayoutTest(testSpec);

    console.log(`Results:`);
    console.log(`  Passed:  ${result.passed}`);
    console.log(`  Failed:  ${result.failed}`);
    console.log(`  Skipped: ${result.skipped}`);
    console.log();

    // Verify all assertions passed
    assert.strictEqual(result.failed, 0, 'All assertions should pass');
    assert.strictEqual(result.passed, 3, 'Expected 3 passing assertions');

    console.log('✓ Test passed!');
    process.exit(0);
  } catch (err) {
    console.error('✗ Test failed:', err.message);
    process.exit(1);
  }
}

main();
