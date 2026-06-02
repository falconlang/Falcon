import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { get } from 'svelte/store';
import {
  __companionInternalsForTests,
  extractComponentDefs,
  generateCompanionCode,
  normalizeCompanionDesignSource,
} from '../src/lib/companion.js';
import { splitFalconSourceByTopLevelLines } from '../src/lib/cell-splitting.js';
import {
  DEBUG_BREAK_PREFIX,
  DEBUG_TRACE_PREFIX,
  DEBUG_VALUE_PREFIX,
  buildFalconSourceMap,
  parseDebugRuntimeEvent,
  parseDebugTraceValue,
} from '../src/lib/debug-source-map.js';
import { ensureDebugNotifierDesignSource, instrumentFalconSourceForDebug } from '../src/lib/debug-instrumentation.js';
import { createQrSvg } from '../src/lib/qr-code.js';
import {
  activeCellId,
  appendDebugLogsFromCompanionResponse,
  cells,
  copiedCellAvailable,
  copyCellById,
  debugLogs,
  getProjectSnapshot,
  loadProjectState,
  pasteCopiedCellBelow,
  projectProperties,
  replaceCodeCells,
  screenList,
  setProjectProperty,
} from '../src/lib/stores.js';
import {
  PROJECT_PROPERTY_DEFINITIONS,
  VISIBLE_PROJECT_PROPERTY_DEFINITIONS,
  applyProjectPropertiesToScmProperties,
  extractProjectPropertiesFromScmProperties,
  normalizeProjectProperties,
  projectPropertiesToAiaProperties,
} from '../src/lib/project-properties.js';
import {
  parseProperties,
  projectPropertiesText,
} from '../src/lib/aia-project-properties.js';
import {
  CURRENT_YA_VERSION,
  componentDefinitionsFromScmProperties,
  formJsonForBlockUpgrade,
  upgradeLegacyScmText,
} from '../src/lib/appinventor-legacy.js';
import {
  buildComponentProps,
  customDesignerPropertyReport,
  unknownDesignerEditorTypes,
} from '../src/lib/designer-properties.js';
import {
  emptyListViewRow,
  listViewColumnsForLayout,
  parseListViewData,
  pruneListViewDataForLayout,
  renameListViewDataAsset,
  serializeListViewData,
} from '../src/lib/listview-data.js';

const simpleComponents = JSON.parse(readFileSync(new URL('../../lang/code/compdb/simple_components.json', import.meta.url), 'utf8'));

test('extractComponentDefs includes anonymous designer components', () => {
  const defs = extractComponentDefs(`
    Screen.Screen1 {
      Label { Text: "First" }
      HorizontalArrangement {
        Button.SaveButton
        Label { Text: "Second" }
      }
    }
  `);

  assert.deepEqual(defs.Screen, ['Screen1']);
  assert.deepEqual(defs.Label, ['Label1', 'Label2']);
  assert.deepEqual(defs.HorizontalArrangement, ['HorizontalArrangement1']);
  assert.deepEqual(defs.Button, ['SaveButton']);
});

test('extractComponentDefs ignores unknown designer component types', () => {
  const defs = extractComponentDefs(`
    Screen.Screen1 {
      Button.SaveButton
      MysteryThing.Hidden
    }
  `);

  assert.deepEqual(defs.Screen, ['Screen1']);
  assert.deepEqual(defs.Button, ['SaveButton']);
  assert.equal(defs.MysteryThing, undefined);
});

test('extractComponentDefs handles same-line bodyless children and current component names', () => {
  const defs = extractComponentDefs('Screen.Screen1 { Button.A, Button.B, NxtSoundSensor.SoundSensor1 }');

  assert.deepEqual(defs.Screen, ['Screen1']);
  assert.deepEqual(defs.Button, ['A', 'B']);
  assert.deepEqual(defs.NxtSoundSensor, ['SoundSensor1']);
  assert.equal(defs.NxtSound, undefined);
});

test('normalizeCompanionDesignSource supplies empty screen fallback and active screen root name', () => {
  assert.equal(
    normalizeCompanionDesignSource('', 'Screen2'),
    'Screen.Screen2 { Title: "Screen2" }'
  );
  assert.equal(
    normalizeCompanionDesignSource('Screen.Screen1 { Title: "Old" }', 'Screen2'),
    'Screen.Screen2 { Title: "Old" }'
  );
});

test('generateCompanionCode returns a 5 character uppercase alphanumeric code', () => {
  for (let i = 0; i < 100; i += 1) {
    assert.match(generateCompanionCode(), /^[A-Z0-9]{5}$/);
  }
});

test('createQrSvg enforces 5 character uppercase alphanumeric companion codes', () => {
  const svg = createQrSvg('AB12C');
  assert.match(svg, /^<svg\b/);
  assert.match(svg, /aria-label="Companion code AB12C"/);

  assert.throws(() => createQrSvg('123456789'), /5 uppercase alphanumeric/);
  assert.throws(() => createQrSvg('abc12'), /5 uppercase alphanumeric/);
});

test('buildFalconSourceMap maps cells to unified editor lines', () => {
  const result = buildFalconSourceMap([
    { id: 'a', type: 'code', code: 'first()\nsecond()' },
    { id: 'm', type: 'markdown', content: 'ignored' },
    { id: 'b', type: 'code', code: 'third()' },
  ]);

  assert.equal(result.source, 'first()\nsecond()\n\nthird()');
  assert.deepEqual(result.entries.map(entry => [entry.cellId, entry.cellLine, entry.unifiedLine]), [
    ['a', 1, 1],
    ['a', 2, 2],
    ['b', 1, 4],
  ]);
});

test('instrumentFalconSourceForDebug emits hidden trace calls for executable lines', () => {
  const source = `func helper() = {
  // ignored
  if (true) {
    println("yes")
  } else {
    println("no")
  }
}`;
  const map = buildFalconSourceMap([{ id: 'c1', type: 'code', code: source }]);
  const result = instrumentFalconSourceForDebug(source, map.entries, {
    sessionId: 'dbg-test',
    notifierName: 'TensorDebugNotifier',
  });

  assert.match(result.source, /TensorDebugNotifier\.LogInfo/);
  assert.equal(result.tracePoints.some(point => point.cellLine === 2), false);
  assert.deepEqual(
    result.tracePoints.map(point => [point.cellId, point.cellLine, point.unifiedLine]),
    [
      ['c1', 3, 3],
      ['c1', 4, 4],
      ['c1', 6, 6],
    ],
  );

  assert.match(result.source, /__TENSOR_DEBUG__\{\\\"type\\\":\\\"trace\\\"/);
  const trace = parseDebugTraceValue({
    type: 'log',
    item: DEBUG_TRACE_PREFIX + JSON.stringify(result.tracePoints[0]),
  });
  assert.equal(trace.sessionId, 'dbg-test');
  assert.equal(trace.cellId, 'c1');
});

test('instrumentFalconSourceForDebug preserves cell-local line numbers across cells', () => {
  const map = buildFalconSourceMap([
    { id: 'a', type: 'code', code: 'when Screen1.Initialize {\n  first()\n}' },
    { id: 'b', type: 'code', code: 'when Button1.Click {\n  second()\n  third()\n}' },
  ]);
  const result = instrumentFalconSourceForDebug(map.source, map.entries, {
    sessionId: 'dbg-cells',
    notifierName: 'TensorDebugNotifier',
  });

  assert.deepEqual(
    result.tracePoints.map(point => [point.cellId, point.cellLine, point.unifiedLine]),
    [
      ['a', 2, 2],
      ['b', 2, 6],
      ['b', 3, 7],
    ],
  );
});

test('instrumentFalconSourceForDebug emits breakpoint probes and expression catalog entries', () => {
  const source = `func checkTextBoxes() = {
  local firstIsNumb = firstNumberTextBox.Text ? number
  local secondIsNumb = secondNumberTextBox.Text ? number
  local eitherIsInvalid = !firstIsNumb || !secondIsNumb
  if (eitherIsInvalid) {
    Notifier1.ShowAlert("Please enter numeric values in the textbox!")
  }
}`;
  const map = buildFalconSourceMap([{ id: 'c1', type: 'code', code: source }]);
  const result = instrumentFalconSourceForDebug(source, map.entries, {
    sessionId: 'dbg-break',
    notifierName: 'TensorDebugNotifier',
    breakpoints: [{ cellId: 'c1', cellLine: 6 }],
  });

  assert.match(result.source, /global tensorDebugContinueFlag_dbg_break = false/);
  assert.match(result.source, /tensorDebugBreak_dbg_break\(/);
  assert.match(result.source, /tensorDebugValue_dbg_break\("dbg-break:expr:expression:4:\d+", !firstIsNumb\)/);
  assert.deepEqual(result.breakpointPoints.map(point => [point.cellId, point.cellLine, point.unifiedLine]), [
    ['c1', 6, 6],
  ]);
  assert.ok(result.expressionCatalog.some(entry => entry.sourceText === 'firstIsNumb'));
  assert.ok(result.expressionCatalog.some(entry => entry.sourceText === '!firstIsNumb'));
});

test('instrumentFalconSourceForDebug avoids user-defined helper identifier collisions', () => {
  const source = `func tensorDebugValue_dbg_collision(a, b) = {
  999
}

func tensorDebugBreak_dbg_collision(payload) {
}

global tensorDebugContinueFlag_dbg_collision = true

func foo() = {
  local x = 1
  x
}`;
  const map = buildFalconSourceMap([{ id: 'c1', type: 'code', code: source }]);
  const result = instrumentFalconSourceForDebug(source, map.entries, {
    sessionId: 'dbg-collision',
    notifierName: 'TensorDebugNotifier',
  });

  assert.deepEqual(result.helpers, {
    continueGlobal: 'tensorDebugContinueFlag_dbg_collision_2',
    valueFunc: 'tensorDebugValue_dbg_collision_2',
    breakFunc: 'tensorDebugBreak_dbg_collision_2',
  });
  assert.match(result.source, /func tensorDebugValue_dbg_collision_2\(/);
  assert.match(result.source, /local x = tensorDebugValue_dbg_collision_2\(/);
  assert.match(result.source, /func tensorDebugValue_dbg_collision\(a, b\)/);
});

test('parseDebugRuntimeEvent handles value captures and breakpoint hits', () => {
  const value = parseDebugRuntimeEvent({
    type: 'log',
    item: `${DEBUG_VALUE_PREFIX}expr-1\ttrue`,
  });
  assert.deepEqual(
    { type: value.type, exprId: value.exprId, value: value.value },
    { type: 'value', exprId: 'expr-1', value: 'true' },
  );

  const hit = parseDebugRuntimeEvent({
    type: 'log',
    item: DEBUG_BREAK_PREFIX + JSON.stringify({
      sessionId: 'dbg',
      hitId: 'hit-1',
      cellId: 'c1',
      cellLine: 3,
      unifiedLine: 7,
    }),
  });
  assert.deepEqual(
    { type: hit.type, sessionId: hit.sessionId, hitId: hit.hitId, cellLine: hit.cellLine, unifiedLine: hit.unifiedLine },
    { type: 'breakpoint-hit', sessionId: 'dbg', hitId: 'hit-1', cellLine: 3, unifiedLine: 7 },
  );
});

test('ensureDebugNotifierDesignSource injects a unique compile-only notifier', () => {
  const injected = ensureDebugNotifierDesignSource('Screen.Screen1 { Title: "Demo" }');

  assert.equal(injected.notifierName, 'TensorDebugNotifier');
  assert.match(injected.designSource, /Title: "Demo",\n  Notifier\.TensorDebugNotifier\n}/);

  const second = ensureDebugNotifierDesignSource(injected.designSource);
  assert.equal(second.notifierName, 'TensorDebugNotifier2');
  assert.match(second.designSource, /Notifier\.TensorDebugNotifier2/);
});

test('appendDebugLogsFromCompanionResponse hides internal debug trace logs', () => {
  const originalLogs = get(debugLogs);
  debugLogs.set([]);

  appendDebugLogsFromCompanionResponse({
    values: [
      { type: 'log', item: '__TENSOR_DEBUG__{"sessionId":"dbg","cellId":"c1","cellLine":1,"unifiedLine":1}' },
      { type: 'log', item: 'visible', level: 'info' },
    ],
  });

  assert.deepEqual(get(debugLogs).map(log => log.message), ['visible']);
  debugLogs.set(originalLogs);
});

test('splitFalconSourceByTopLevelLines creates cells from parser line starts', () => {
  const source = `// prelude

func helper() = {
  1
}

when Button1.Click {
  helper()
}`;

  assert.deepEqual(splitFalconSourceByTopLevelLines(source, [3, 7]), [
    '// prelude\n\nfunc helper() = {\n  1\n}',
    'when Button1.Click {\n  helper()\n}',
  ]);
});

test('splitFalconSourceByTopLevelLines groups globals into the first cell', () => {
  const source = `// score state
global score = 0

when Button1.Click {
  globalScore()
}

// player state
global player = "Ada"

func globalScore() = score`;

  assert.deepEqual(splitFalconSourceByTopLevelLines(source, [2, 4, 9, 11]), [
    '// score state\nglobal score = 0\n\n// player state\nglobal player = "Ada"',
    'when Button1.Click {\n  globalScore()\n}',
    'func globalScore() = score',
  ]);
});

test('splitFalconSourceByTopLevelLines infers top-level cells without parser line starts', () => {
  const source = `global names = [
  "Ada",
  "Grace"
]

func greet() = {
  println("hi")
}

when Button1.Click {
  greet()
}`;

  assert.deepEqual(splitFalconSourceByTopLevelLines(source), [
    'global names = [\n  "Ada",\n  "Grace"\n]',
    'func greet() = {\n  println("hi")\n}',
    'when Button1.Click {\n  greet()\n}',
  ]);
});

test('replaceCodeCells preserves markdown and rebuilds code cells', () => {
  const originalCells = get(cells);
  const originalActive = get(activeCellId);

  cells.set([
    { id: 'm1', type: 'markdown', content: 'Intro' },
    { id: 'old1', type: 'code', code: 'old one', execCount: 7 },
    { id: 'old2', type: 'code', code: 'old two', execCount: null },
    { id: 'm2', type: 'markdown', content: 'Outro' },
  ]);

  replaceCodeCells(['first()', 'second()', 'third()'], { activeIndex: 1 });
  const nextCells = get(cells);
  assert.deepEqual(nextCells.map(cell => [cell.id, cell.type, cell.code ?? cell.content]), [
    ['m1', 'markdown', 'Intro'],
    ['old1', 'code', 'first()'],
    ['old2', 'code', 'second()'],
    [nextCells[3].id, 'code', 'third()'],
    ['m2', 'markdown', 'Outro'],
  ]);
  assert.equal(get(activeCellId), 'old2');

  cells.set(originalCells);
  activeCellId.set(originalActive);
});

test('replaceCodeCells preserves interleaved markdown anchors', () => {
  const originalCells = get(cells);
  const originalActive = get(activeCellId);

  cells.set([
    { id: 'old1', type: 'code', code: 'old one', execCount: 7 },
    { id: 'm1', type: 'markdown', content: 'Between' },
    { id: 'old2', type: 'code', code: 'old two', execCount: null },
    { id: 'm2', type: 'markdown', content: 'Outro' },
  ]);

  replaceCodeCells(['first()', 'second()', 'third()']);
  const nextCells = get(cells);
  assert.deepEqual(nextCells.map(cell => [cell.id, cell.type, cell.code ?? cell.content]), [
    ['old1', 'code', 'first()'],
    ['m1', 'markdown', 'Between'],
    ['old2', 'code', 'second()'],
    [nextCells[3].id, 'code', 'third()'],
    ['m2', 'markdown', 'Outro'],
  ]);

  cells.set(originalCells);
  activeCellId.set(originalActive);
});

test('loadProjectState makes duplicate screen names unique', () => {
  const originalProject = getProjectSnapshot();

  loadProjectState({
    projectName: 'DupScreens',
    activeScreen: 'Screen1',
    screens: [
      { name: 'Screen1', cells: [], designCode: '' },
      { name: 'Screen1', cells: [], designCode: '' },
      { name: 'Screen2', cells: [], designCode: '' },
    ],
  });

  assert.deepEqual(get(screenList), ['Screen1', 'Screen2', 'Screen3']);

  loadProjectState(originalProject);
});

test('copyCellById and pasteCopiedCellBelow clone a cell below the target', () => {
  const originalCells = get(cells);
  const originalFirst = originalCells[0];
  const originalSecond = originalCells[1];

  assert.equal(copyCellById(originalFirst.id), true);
  assert.equal(get(copiedCellAvailable), true);

  const pastedId = pasteCopiedCellBelow(originalSecond.id);
  const nextCells = get(cells);
  const pastedIndex = nextCells.findIndex(cell => cell.id === pastedId);

  assert.notEqual(pastedId, originalFirst.id);
  assert.equal(pastedIndex, 2);
  assert.equal(nextCells[pastedIndex].type, originalFirst.type);
  assert.equal(nextCells[pastedIndex].code, originalFirst.code);

  cells.set(originalCells);
});

test('project property definitions include App Inventor project properties and hide BlocksToolkit from UI', () => {
  assert.equal(PROJECT_PROPERTY_DEFINITIONS.length, 20);
  assert.ok(PROJECT_PROPERTY_DEFINITIONS.some(property => property.name === 'BlocksToolkit'));
  assert.equal(VISIBLE_PROJECT_PROPERTY_DEFINITIONS.some(property => property.name === 'BlocksToolkit'), false);
  assert.deepEqual(
    PROJECT_PROPERTY_DEFINITIONS.filter(property => property.category === 'iOS Settings').map(property => property.name),
    [
      'NSBluetoothAlwaysUsageDescription',
      'NSBluetoothPeripheralUsageDescription',
      'NSContactsUsageDescription',
      'NSMicrophoneUsageDescription',
      'NSCameraUsageDescription',
      'NSSpeechRecognitionUsageDescription',
      'NSLocationWhenInUseUsageDescription',
    ]
  );
});

test('project properties normalize AIA keys and serialize back to project.properties keys', () => {
  const normalized = normalizeProjectProperties({
    aname: 'Demo App',
    subsetjson: '{"shownComponentTypes":["Button"]}',
    defaultfilescope: 'Shared',
    showlistsasjson: 'false',
    'color.primary': '#112233',
    NSCameraUsageDescription: 'Take photos',
    customflag: 'kept',
  });

  assert.equal(normalized.AppName, 'Demo App');
  assert.equal(normalized.BlocksToolkit, '{"shownComponentTypes":["Button"]}');
  assert.equal(normalized.DefaultFileScope, 'Shared');
  assert.equal(normalized.ShowListsAsJson, 'False');
  assert.equal(normalized.PrimaryColor, '&HFF112233');

  const aiaProperties = projectPropertiesToAiaProperties(normalized);
  assert.equal(aiaProperties.aname, 'Demo App');
  assert.equal(aiaProperties.subsetjson, '{"shownComponentTypes":["Button"]}');
  assert.equal(aiaProperties.defaultfilescope, 'Shared');
  assert.equal(aiaProperties.showlistsasjson, 'False');
  assert.equal(aiaProperties['color.primary'], '&HFF112233');
  assert.equal(aiaProperties.NSCameraUsageDescription, 'Take photos');
  assert.equal(aiaProperties.customflag, 'kept');
  assert.equal('BlocksToolkit' in aiaProperties, false);
});

test('project property defaults match App Inventor project.properties defaults', () => {
  const normalized = normalizeProjectProperties({});

  assert.equal(normalized.PrimaryColor, '&HFF3F51B5');
  assert.equal(normalized.PrimaryColorDark, '&HFF303F9F');
  assert.equal(normalized.AccentColor, '&HFFFF4081');
});

test('AIA project.properties parser and writer follow Java properties escaping', () => {
  const parsed = parseProperties([
    'aname=Caf\\u00e9',
    'description=first\\',
    '  second',
    'spaced\\ key\\ =\\ leading',
    'path=a\\:b',
  ].join('\n'));

  assert.equal(parsed.aname, 'Café');
  assert.equal(parsed.description, 'firstsecond');
  assert.equal(parsed['spaced key '], ' leading');
  assert.equal(parsed.path, 'a:b');

  const text = projectPropertiesText('RenamedProject', {
    main: 'appinventor.ai_original.OldProject.Screen1',
    aname: 'Café',
  });

  assert.match(text, /^main=appinventor\.ai_original\.RenamedProject\.Screen1$/m);
  assert.match(text, /^aname=Caf\\u00e9$/m);
  assert.match(text, /^color\.primary=&HFF3F51B5$/m);
  assert.match(text, /^color\.primary\.dark=&HFF303F9F$/m);
  assert.match(text, /^color\.accent=&HFFFF4081$/m);
});

test('legacy App Inventor SCM is upgraded to the current Tensor import shape', () => {
  const legacyScm = `#|
$Properties
$Source $Form
$Define Screen1 $As Form
Layout = 1
Layout.Orientation = 1
Title = "Screen1"
$End $Define
$End $Properties

|#
#|
$JSON
{"Source":"Form","Properties":{"$Name":"Screen1","$Type":"Form","Layout":"1","Layout.Orientation":"1","Title":"\\"Screen1\\""}}
|#`;

  const upgraded = upgradeLegacyScmText(legacyScm, { host: 'tensor.test' });
  const props = upgraded.scm.Properties;

  assert.equal(upgraded.scm.YaVersion, CURRENT_YA_VERSION);
  assert.equal(props.$Version, '31');
  assert.equal(props.Uuid, '0');
  assert.equal(props.Title, 'Screen1');
  assert.equal(props.Sizing, 'Fixed');
  assert.equal(props.ShowListsAsJson, 'False');
  assert.equal(props.Theme, 'Classic');
  assert.equal('Layout' in props, false);
  assert.equal('Layout.Orientation' in props, false);
  assert.deepEqual(upgraded.scm.authURL, ['tensor.test']);
  assert.deepEqual(componentDefinitionsFromScmProperties(props), { Screen: ['Screen1'] });

  const blockUpgradeJson = JSON.parse(formJsonForBlockUpgrade(upgraded.original));
  assert.equal(blockUpgradeJson.Properties.$Version, '1');

  const normalized = normalizeProjectProperties(extractProjectPropertiesFromScmProperties(props));
  assert.equal(normalized.Sizing, 'Fixed');
  assert.equal(normalized.ShowListsAsJson, 'False');
  assert.equal(normalized.Theme, 'Classic');
});

test('legacy SCM upgrader applies App Inventor component renames and property migrations', () => {
  const upgraded = upgradeLegacyScmText(JSON.stringify({
    YaVersion: '1',
    Source: 'Form',
    Properties: {
      $Name: 'Screen1',
      $Type: 'Form',
      $Version: '1',
      $Components: [
        { $Name: 'Logger1', $Type: 'Logger', $Version: '1' },
        { $Name: 'Button1', $Type: 'Button', $Version: '1', Alignment: '1' },
        { $Name: 'TextBox1', $Type: 'TextBox', $Version: '3' },
        { $Name: 'File1', $Type: 'File', $Version: '3', LegacyMode: 'True' },
      ],
    },
  }));

  const children = upgraded.scm.Properties.$Components;
  assert.equal(children[0].$Type, 'Notifier');
  assert.equal(children[0].$Version, '6');
  assert.equal(children[1].TextAlignment, '1');
  assert.equal('Alignment' in children[1], false);
  assert.equal(children[2].MultiLine, 'True');
  assert.equal(children[3].DefaultScope, 'Legacy');
  assert.equal('LegacyMode' in children[3], false);
});

test('legacy SCM upgrader keeps unhandled component upgrades marked for App Inventor', () => {
  const upgraded = upgradeLegacyScmText(JSON.stringify({
    YaVersion: '10',
    Source: 'Form',
    Properties: {
      $Name: 'Screen1',
      $Type: 'Form',
      $Version: '31',
      $Components: [
        {
          $Name: 'Starter1',
          $Type: 'ActivityStarter',
          $Version: '1',
          LegacyCustomProperty: 'keep-me',
        },
      ],
    },
  }));

  const child = upgraded.scm.Properties.$Components[0];
  assert.notEqual(upgraded.scm.YaVersion, CURRENT_YA_VERSION);
  assert.equal(upgraded.needsAppInventorUpgrade, true);
  assert.equal(child.$Version, '1');
  assert.equal(child.LegacyCustomProperty, 'keep-me');
  assert.deepEqual(upgraded.warnings.map(warning => [warning.type, warning.name]), [
    ['ActivityStarter', 'Starter1'],
  ]);
});

test('companion event validation reads mutation attributes independent of order', () => {
  const xml = `<xml>
    <block type="component_event">
      <mutation event_name="Click" component_type="Button" instance_name="Button1"></mutation>
    </block>
  </xml>`;

  const defs = __companionInternalsForTests.extractEventDefs(xml);
  assert.deepEqual(defs, [{ component: 'Button1', event: 'Click' }]);
  assert.deepEqual(
    __companionInternalsForTests.findMissingEvents(defs, '(define-event Button1 Click (lambda () #!null))'),
    []
  );
});

test('companion screen validation does not accept substring matches', () => {
  assert.deepEqual(
    __companionInternalsForTests.findMissingScreens({ Screen: ['Screen1'] }, '(define-form Screen10)'),
    ['Screen1']
  );
  assert.deepEqual(
    __companionInternalsForTests.findMissingScreens({ Screen: ['Screen1'] }, '(define-form Screen1)'),
    []
  );
});

test('setProjectProperty updates canonical store values including hidden BlocksToolkit backend', () => {
  const original = get(projectProperties);

  assert.equal(setProjectProperty('AppName', 'Store Demo'), 'Store Demo');
  assert.equal(setProjectProperty('BlocksToolkit', '{"level":"beginner"}'), '{"level":"beginner"}');

  const next = get(projectProperties);
  assert.equal(next.AppName, 'Store Demo');
  assert.equal(next.BlocksToolkit, '{"level":"beginner"}');

  projectProperties.set(original);
});

test('project properties are applied to Screen1 SCM properties', () => {
  const scmProperties = applyProjectPropertiesToScmProperties({
    $Name: 'Screen1',
    $Type: 'Form',
    Title: 'Screen1',
    AppName: 'Old',
    Icon: 'old.png',
    aname: 'legacy',
  }, {
    AppName: '',
    DefaultFileScope: 'Cache',
    BlocksToolkit: '{"toolkit":"all"}',
    Theme: 'Classic',
  }, 'DemoProject');

  assert.equal(scmProperties.$Name, 'Screen1');
  assert.equal(scmProperties.AppName, '');
  assert.equal(scmProperties.DefaultFileScope, 'Cache');
  assert.equal(scmProperties.BlocksToolkit, '{"toolkit":"all"}');
  assert.equal(scmProperties.Theme, 'Classic');
  assert.equal(scmProperties.ActionBar, 'False');
  assert.equal('Icon' in scmProperties, false);
  assert.equal('aname' in scmProperties, false);
});

test('designer property metadata exposes ListView custom designer properties', () => {
  const props = buildComponentProps(simpleComponents);
  const listViewProps = props.ListView;
  const listData = listViewProps.find(prop => prop.name === 'ListData');
  const layout = listViewProps.find(prop => prop.name === 'ListViewLayout');

  assert.equal(listData.editorType, 'ListViewAddData');
  assert.equal(listData.designerOnly, true);
  assert.equal(listData.customEditor.kind, 'dialog');
  assert.equal(listData.customEditor.layoutProperty, 'ListViewLayout');
  assert.deepEqual(layout.options.map(option => option.value), ['0', '1', '2', '3', '4', '5']);

  assert.equal(props.Form.some(prop => prop.name === 'BlocksToolkit'), false);
  assert.equal(props.Form.some(prop => prop.name === 'AppName'), false);
});

test('custom designer property scan currently finds only ListView.ListData dialog editor', () => {
  assert.deepEqual(customDesignerPropertyReport(simpleComponents), [
    {
      component: 'ListView',
      property: 'ListData',
      editorType: 'ListViewAddData',
      kind: 'dialog',
    },
  ]);
  assert.equal(
    unknownDesignerEditorTypes(simpleComponents).some(entry => entry.editorType === 'ListViewAddData'),
    false
  );
});

test('ListView data helpers preserve App Inventor row JSON by layout', () => {
  assert.deepEqual(listViewColumnsForLayout('4'), ['Text1', 'Text2', 'Image']);
  assert.deepEqual(emptyListViewRow('3'), { Text1: '', Image: 'None' });

  const rows = parseListViewData('[{"Text1":"Alpha","Text2":"Detail","Image":"one.png"}]', '4');
  assert.deepEqual(rows, [{ Text1: 'Alpha', Text2: 'Detail', Image: 'one.png' }]);
  assert.equal(serializeListViewData(rows, '1'), '[{"Text1":"Alpha","Text2":"Detail"}]');
  assert.equal(pruneListViewDataForLayout(JSON.stringify(rows), '0'), '[{"Text1":"Alpha"}]');
  assert.equal(
    renameListViewDataAsset(JSON.stringify(rows), 'one.png', 'two.png'),
    '[{"Text1":"Alpha","Text2":"Detail","Image":"two.png"}]'
  );
});
