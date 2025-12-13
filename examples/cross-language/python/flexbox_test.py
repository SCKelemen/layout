#!/usr/bin/env python3
"""
Python example: Testing layout with wptest eval

This demonstrates how to use wptest from Python without any dependencies
beyond the Python standard library. Just spawn wptest and pass JSON via stdin.
"""

import json
import subprocess
import sys

# Test spec - same format works in ANY language
test_spec = {
    "layout": {
        "type": "container",
        "style": {
            "display": "flex",
            "justifyContent": "space-between",
            "alignItems": "center",
            "width": 600,
            "height": 100,
        },
        "children": [
            {"type": "container", "style": {"width": 100, "height": 50}},
            {"type": "container", "style": {"width": 100, "height": 50}},
            {"type": "container", "style": {"width": 100, "height": 50}},
        ],
    },
    "constraints": {
        "maxWidth": 800,
        "maxHeight": 600,
    },
    "assertions": [
        {
            "type": "layout",
            "expression": "getX(child(root(), 0)) == 0.0",
            "message": "first-child-at-start",
        },
        {
            "type": "layout",
            "expression": "getRight(child(root(), 2)) == getWidth(root())",
            "message": "last-child-at-end",
        },
        {
            "type": "layout",
            "expression": "getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0",
            "message": "vertically-centered",
        },
    ],
    "binding": "old",
}


def run_layout_test(spec):
    """
    Run a layout test by spawning wptest and communicating via stdin/stdout.

    Args:
        spec: Test specification dictionary

    Returns:
        dict: Test results with passed, failed, skipped counts

    Raises:
        RuntimeError: If wptest fails or returns invalid JSON
    """
    try:
        # Spawn wptest eval command
        result = subprocess.run(
            ["wptest", "eval"],
            input=json.dumps(spec).encode("utf-8"),
            capture_output=True,
            check=True,
        )

        # Parse JSON output
        return json.loads(result.stdout)

    except subprocess.CalledProcessError as err:
        raise RuntimeError(
            f"wptest exited with code {err.returncode}\n{err.stderr.decode()}"
        )
    except json.JSONDecodeError as err:
        raise RuntimeError(
            f"Failed to parse wptest output: {err}\n{result.stdout.decode()}"
        )
    except FileNotFoundError:
        raise RuntimeError(
            "wptest not found - ensure it's built and in your PATH"
        )


def main():
    """Run the flexbox layout test."""
    print("Running flexbox layout test from Python...\n")

    try:
        result = run_layout_test(test_spec)

        print(f"Results:")
        print(f"  Passed:  {result['passed']}")
        print(f"  Failed:  {result['failed']}")
        print(f"  Skipped: {result['skipped']}")
        print()

        # Verify all assertions passed
        assert result["failed"] == 0, "All assertions should pass"
        assert result["passed"] == 3, "Expected 3 passing assertions"

        print("✓ Test passed!")
        sys.exit(0)

    except (RuntimeError, AssertionError) as err:
        print(f"✗ Test failed: {err}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
