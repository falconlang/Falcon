import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  blocklyTargetXmlWithContextParts,
  blocklyTargetXmlsWithContextParts,
  topLevelIndexesFullyContainedInLineRange,
} from '../src/lib/blockly-preview-selection.js';

const xml = [
  '<block type="first"></block>',
  '<block type="second"></block>',
  '<block type="third"></block>',
].join('\0');

test('blocklyTargetXmlWithContextParts selects one target and keeps other chunks as context', () => {
  const result = blocklyTargetXmlWithContextParts(xml, 1);

  assert.equal(result.targetXml, '<block type="second"></block>');
  assert.equal(
    result.contextXml,
    '<block type="first"></block>\0<block type="third"></block>',
  );
  assert.equal(result.targetIndex, 1);
});

test('blocklyTargetXmlWithContextParts defaults to the last XML chunk', () => {
  const result = blocklyTargetXmlWithContextParts(xml);

  assert.equal(result.targetXml, '<block type="third"></block>');
  assert.equal(result.contextXml, '<block type="first"></block>\0<block type="second"></block>');
  assert.equal(result.targetIndex, 2);
});

test('blocklyTargetXmlsWithContextParts selects multiple targets in source order', () => {
  const result = blocklyTargetXmlsWithContextParts(xml, [2, 0]);

  assert.equal(
    result.targetXml,
    '<block type="first"></block>\0<block type="third"></block>',
  );
  assert.equal(result.contextXml, '<block type="second"></block>');
  assert.deepEqual(result.targetIndexes, [0, 2]);
});

test('topLevelIndexesFullyContainedInLineRange returns every fully selected expression', () => {
  const source = [
    'global a = 1',
    'global b = 2',
    'func c() = {',
    '  a + b',
    '}',
    'global d = 4',
  ].join('\n');

  assert.deepEqual(
    topLevelIndexesFullyContainedInLineRange(source, [1, 2, 3, 6], 2, 5),
    [1, 2],
  );
});

test('topLevelIndexesFullyContainedInLineRange excludes partial expressions', () => {
  const source = [
    'global a = 1',
    'func b() = {',
    '  a + 1',
    '}',
    'global c = 3',
  ].join('\n');

  assert.deepEqual(
    topLevelIndexesFullyContainedInLineRange(source, [1, 2, 5], 1, 3),
    [0],
  );
});

test('topLevelIndexesFullyContainedInLineRange does not require trailing blank lines', () => {
  const source = [
    'global a = 1',
    '',
    'global b = 2',
  ].join('\n');

  assert.deepEqual(
    topLevelIndexesFullyContainedInLineRange(source, [1, 3], 1, 1),
    [0],
  );
});

test('topLevelIndexesFullyContainedInLineRange ignores blank and comment-only selections', () => {
  const source = [
    'global a = 1',
    '',
    '// comment',
    'global b = 2',
  ].join('\n');

  assert.deepEqual(
    topLevelIndexesFullyContainedInLineRange(source, [1, 4], 2, 3),
    [],
  );
});
