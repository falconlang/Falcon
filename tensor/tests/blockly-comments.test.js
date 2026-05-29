import assert from 'node:assert/strict';
import { test } from 'node:test';
import { leadingFalconCommentBeforeLine } from '../src/lib/blockly-comments.js';

test('leadingFalconCommentBeforeLine collects adjacent Falcon line comments', () => {
  const source = `// First line
  // second line
func check() = true`;

  assert.equal(
    leadingFalconCommentBeforeLine(source, 3),
    'First line\nsecond line'
  );
});

test('leadingFalconCommentBeforeLine does not cross blank lines', () => {
  const source = `// detached

func check() = true`;

  assert.equal(leadingFalconCommentBeforeLine(source, 3), '');
});
