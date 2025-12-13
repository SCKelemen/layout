#!/usr/bin/env rust-script
//! ```cargo
//! [dependencies]
//! serde_json = "1.0"
//! serde = { version = "1.0", features = ["derive"] }
//! ```
//!
//! Rust example: Testing layout with wptest eval
//!
//! This demonstrates how to use wptest from Rust without any custom bindings.
//! Just spawn wptest and pass JSON via stdin. Uses serde_json for JSON handling.
//!
//! To run:
//!   cargo install rust-script
//!   rust-script flexbox_test.rs
//!
//! Or compile normally:
//!   rustc flexbox_test.rs -o flexbox_test
//!   ./flexbox_test

use serde::{Deserialize, Serialize};
use std::io::Write;
use std::process::{Command, Stdio};

#[derive(Serialize, Deserialize, Debug)]
struct TestSpec {
    layout: Layout,
    constraints: Constraints,
    assertions: Vec<Assertion>,
    binding: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct Layout {
    #[serde(rename = "type")]
    layout_type: String,
    style: Style,
    #[serde(skip_serializing_if = "Option::is_none")]
    children: Option<Vec<Layout>>,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
struct Style {
    #[serde(skip_serializing_if = "Option::is_none")]
    display: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    justify_content: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    align_items: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    width: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    height: Option<f64>,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
struct Constraints {
    #[serde(skip_serializing_if = "Option::is_none")]
    max_width: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    max_height: Option<f64>,
}

#[derive(Serialize, Deserialize, Debug)]
struct Assertion {
    #[serde(rename = "type")]
    assertion_type: String,
    expression: String,
    message: String,
}

#[derive(Deserialize, Debug)]
struct TestResult {
    passed: usize,
    failed: usize,
    skipped: usize,
}

fn run_layout_test(spec: &TestSpec) -> Result<TestResult, Box<dyn std::error::Error>> {
    // Serialize test spec to JSON
    let json_input = serde_json::to_string(spec)?;

    // Spawn wptest eval
    let mut child = Command::new("wptest")
        .arg("eval")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("Failed to spawn wptest: {}", e))?;

    // Write JSON to stdin
    if let Some(mut stdin) = child.stdin.take() {
        stdin
            .write_all(json_input.as_bytes())
            .map_err(|e| format!("Failed to write to stdin: {}", e))?;
    }

    // Wait for completion and collect output
    let output = child
        .wait_with_output()
        .map_err(|e| format!("Failed to wait for wptest: {}", e))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(format!("wptest failed: {}", stderr).into());
    }

    // Parse JSON output
    let result: TestResult = serde_json::from_slice(&output.stdout)
        .map_err(|e| format!("Failed to parse JSON: {}", e))?;

    Ok(result)
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("Running flexbox layout test from Rust...\n");

    // Build test spec
    let test_spec = TestSpec {
        layout: Layout {
            layout_type: "container".to_string(),
            style: Style {
                display: Some("flex".to_string()),
                justify_content: Some("space-between".to_string()),
                align_items: Some("center".to_string()),
                width: Some(600.0),
                height: Some(100.0),
            },
            children: Some(vec![
                Layout {
                    layout_type: "container".to_string(),
                    style: Style {
                        display: None,
                        justify_content: None,
                        align_items: None,
                        width: Some(100.0),
                        height: Some(50.0),
                    },
                    children: None,
                },
                Layout {
                    layout_type: "container".to_string(),
                    style: Style {
                        display: None,
                        justify_content: None,
                        align_items: None,
                        width: Some(100.0),
                        height: Some(50.0),
                    },
                    children: None,
                },
                Layout {
                    layout_type: "container".to_string(),
                    style: Style {
                        display: None,
                        justify_content: None,
                        align_items: None,
                        width: Some(100.0),
                        height: Some(50.0),
                    },
                    children: None,
                },
            ]),
        },
        constraints: Constraints {
            max_width: Some(800.0),
            max_height: Some(600.0),
        },
        assertions: vec![
            Assertion {
                assertion_type: "layout".to_string(),
                expression: "getX(child(root(), 0)) == 0.0".to_string(),
                message: "first-child-at-start".to_string(),
            },
            Assertion {
                assertion_type: "layout".to_string(),
                expression: "getRight(child(root(), 2)) == getWidth(root())".to_string(),
                message: "last-child-at-end".to_string(),
            },
            Assertion {
                assertion_type: "layout".to_string(),
                expression: "getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0".to_string(),
                message: "vertically-centered".to_string(),
            },
        ],
        binding: "old".to_string(),
    };

    // Run the test
    let result = run_layout_test(&test_spec)?;

    println!("Results:");
    println!("  Passed:  {}", result.passed);
    println!("  Failed:  {}", result.failed);
    println!("  Skipped: {}", result.skipped);
    println!();

    // Verify all assertions passed
    if result.failed > 0 {
        eprintln!("✗ Test failed: {} assertions failed", result.failed);
        std::process::exit(1);
    }

    if result.passed != 3 {
        eprintln!("✗ Test failed: Expected 3 passing assertions, got {}", result.passed);
        std::process::exit(1);
    }

    println!("✓ Test passed!");
    Ok(())
}
