export const SIMULATION_SUPPORTED_TYPES = new Set([
  'Screen',
  'Form',
  'Label',
  'TextBox',
  'PasswordTextBox',
  'Button',
  'CheckBox',
  'Switch',
  'Slider',
  'Image',
  'Spinner',
  'ListPicker',
  'DatePicker',
  'TimePicker',
  'ListView',
  'HorizontalArrangement',
  'VerticalArrangement',
  'HorizontalScrollArrangement',
  'VerticalScrollArrangement',
  'AbsoluteArrangement',
  'Notifier',
  'TinyDB',
]);

export const SIMULATION_NONVISIBLE_TYPES = new Set([
  'Notifier',
  'TinyDB',
]);

const COMMON_VISIBLE_PROPS = {
  Visible: true,
  Enabled: true,
  Width: -1,
  Height: -1,
};

const VIEWPORT_DEFAULTS = {
  Width: 360,
  Height: 640,
};

const COMMON_FONT_PROPS = {
  FontSize: 14,
  FontBold: false,
  FontItalic: false,
  FontTypeface: '0',
};

const COMMON_TEXT_PROPS = {
  ...COMMON_FONT_PROPS,
  TextColor: '&HFF000000',
  TextAlignment: 0,
};

const BUTTON_STYLE_PROPS = {
  ...COMMON_FONT_PROPS,
  BackgroundColor: '',
  TextColor: '&HFFFFFFFF',
  TextAlignment: 1,
  Image: '',
  Shape: 0,
  ShowFeedback: true,
};

const ARRANGEMENT_PROPS = {
  Visible: true,
  AlignHorizontal: 1,
  AlignVertical: 1,
  BackgroundColor: 'transparent',
  Image: '',
};

const MONTH_NAMES = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

function dateInstant(year, month, day) {
  return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}T00:00:00`;
}

function timeInstant(hour, minute) {
  return `1970-01-01T${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}:00`;
}

export function buildSimulationDefaults(now = new Date()) {
  const current = now instanceof Date && Number.isFinite(now.getTime()) ? now : new Date();
  const year = current.getFullYear();
  const month = current.getMonth() + 1;
  const day = current.getDate();
  const hour = current.getHours();
  const minute = current.getMinutes();

  return {
    Screen: {
      Visible: true,
      Enabled: true,
      ...VIEWPORT_DEFAULTS,
      AlignHorizontal: 1,
      AlignVertical: 1,
      Scrollable: false,
      BackgroundColor: '&HFFFFFFFF',
      BackgroundImage: '',
      Title: '',
      TitleVisible: true,
      ScreenOrientation: 'unspecified',
      ShowStatusBar: true,
      Platform: 'Android',
    },
    Form: {
      Visible: true,
      Enabled: true,
      ...VIEWPORT_DEFAULTS,
      AlignHorizontal: 1,
      AlignVertical: 1,
      Scrollable: false,
      BackgroundColor: '&HFFFFFFFF',
      BackgroundImage: '',
      Title: '',
      TitleVisible: true,
      ScreenOrientation: 'unspecified',
      ShowStatusBar: true,
      Platform: 'Android',
    },
    Label: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      BackgroundColor: 'transparent',
      HasMargins: true,
      HTMLFormat: false,
    },
    TextBox: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      Hint: '',
      BackgroundColor: '&HFFFFFFFF',
      HintColor: '&HFFAAAAAA',
      NumbersOnly: false,
      MultiLine: false,
      ReadOnly: false,
      Width: 160,
    },
    PasswordTextBox: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      Hint: '',
      BackgroundColor: '&HFFFFFFFF',
      HintColor: '&HFFAAAAAA',
      NumbersOnly: false,
      PasswordVisible: false,
      ReadOnly: false,
      Width: 160,
    },
    Button: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
    },
    CheckBox: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      Checked: false,
      BackgroundColor: 'transparent',
    },
    Switch: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      On: false,
      BackgroundColor: 'transparent',
      ThumbColorActive: '&HFFFFFFFF',
      ThumbColorInactive: '&HFFCCCCCC',
      TrackColorActive: '&HFF00FF00',
      TrackColorInactive: '&HFF444444',
    },
    Slider: {
      Visible: true,
      Enabled: true,
      Width: -1,
      Height: -1,
      MinValue: 10,
      MaxValue: 50,
      ThumbPosition: 30,
      ThumbEnabled: true,
      ColorLeft: '&HFFFFC800',
      ColorRight: '&HFF888888',
      ThumbColor: '&HFF444444',
      NumberOfSteps: 100,
    },
    Image: {
      Visible: true,
      Width: -1,
      Height: -1,
      Picture: '',
      AlternateText: '',
      Clickable: false,
      RotationAngle: 0,
      ScalePictureToFit: false,
      Scaling: 0,
      Animation: '',
    },
    Spinner: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_FONT_PROPS,
      TextColor: '&HFF000000',
      TextAlignment: 1,
      BackgroundColor: '',
      Elements: [],
      ElementsFromString: '',
      Selection: '',
      SelectionIndex: 0,
      Prompt: '',
    },
    ListPicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
      Elements: [],
      ElementsFromString: '',
      Selection: '',
      SelectionIndex: 0,
      Title: '',
      ShowFilterBar: false,
      ItemTextColor: '&HFFFFFFFF',
      ItemBackgroundColor: '&HFF000000',
    },
    DatePicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: 'Set Date',
      BackgroundColor: '&HFFFFFFFF',
      TextColor: '&HFF000000',
      Day: day,
      Month: month,
      Year: year,
      MonthInText: MONTH_NAMES[month - 1] || '',
      Instant: dateInstant(year, month, day),
    },
    TimePicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: 'Set Time',
      BackgroundColor: '&HFFFFFFFF',
      TextColor: '&HFF000000',
      Hour: hour,
      Minute: minute,
      Instant: timeInstant(hour, minute),
    },
    ListView: {
      Visible: true,
      Enabled: true,
      Width: -1,
      Height: -2,
      Elements: [],
      ElementsFromString: '',
      Selection: '',
      SelectionIndex: 0,
      BackgroundColor: '&HFF000000',
      TextColor: '&HFFFFFFFF',
      FontSize: 22,
      FontSizeDetail: 14,
      FontTypeface: '0',
      FontTypefaceDetail: '0',
      TextColorDetail: '&HFFFFFFFF',
      SelectionColor: '&HFFD3D3D3',
      ShowFilterBar: false,
      HintText: 'Search list...',
      ListData: '',
      ListViewLayout: 0,
      Orientation: 1,
      ImageWidth: 200,
      ImageHeight: 200,
      ElementColor: 'transparent',
      DividerColor: '&HFFFFFFFF',
      DividerThickness: 0,
      ElementMarginsWidth: 0,
      ElementCornerRadius: 0,
      BounceEdgeEffect: false,
      MultiSelect: false,
    },
    HorizontalArrangement: { ...ARRANGEMENT_PROPS, Width: -1, Height: -1 },
    VerticalArrangement: { ...ARRANGEMENT_PROPS, Width: -1, Height: -1 },
    HorizontalScrollArrangement: { ...ARRANGEMENT_PROPS, Width: -1, Height: -1 },
    VerticalScrollArrangement: { ...ARRANGEMENT_PROPS, Width: -1, Height: -1 },
    AbsoluteArrangement: { Visible: true, Width: -2, Height: 100, BackgroundColor: '&HFFFFFFFF', Image: '' },
    Notifier: { Visible: false, NotifierLength: 1, BackgroundColor: '&HFF444444', TextColor: '&HFFFFFFFF' },
    TinyDB: { Visible: false, Namespace: 'TinyDB1' },
  };
}

export const SIMULATION_DEFAULTS = buildSimulationDefaults();

export const SIMULATION_BEHAVIOR_PROPS = new Set([
  'Namespace',
]);

const SIMULATION_IGNORED_PROPS_BY_TYPE = {
  Image: new Set(['Enabled']),
};

export const SIMULATION_VISUAL_PROPS = new Set([
  'Text',
  'Hint',
  'HintColor',
  'Visible',
  'Enabled',
  'Checked',
  'On',
  'NumbersOnly',
  'PasswordVisible',
  'MultiLine',
  'ReadOnly',
  'Width',
  'Height',
  'Left',
  'Top',
  'BackgroundColor',
  'BackgroundImage',
  'TextColor',
  'HintText',
  'FontSize',
  'FontSizeDetail',
  'FontBold',
  'FontItalic',
  'FontTypeface',
  'FontTypefaceDetail',
  'TextAlignment',
  'HasMargins',
  'HTMLFormat',
  'HTMLContent',
  'Title',
  'TitleVisible',
  'AboutScreen',
  'ActionBar',
  'PrimaryColor',
  'PrimaryColorDark',
  'AccentColor',
  'HighContrast',
  'BigDefaultText',
  'Sizing',
  'OpenScreenAnimation',
  'CloseScreenAnimation',
  'Theme',
  'ScreenOrientation',
  'ShowStatusBar',
  'Platform',
  'PlatformVersion',
  'Scrollable',
  'AlignHorizontal',
  'AlignVertical',
  'Picture',
  'Image',
  'Clickable',
  'RotationAngle',
  'ScalePictureToFit',
  'Scaling',
  'AlternateText',
  'Animation',
  'Shape',
  'ShowFeedback',
  'Elements',
  'ElementsFromString',
  'Selection',
  'SelectionIndex',
  'Prompt',
  'ShowFilterBar',
  'ItemTextColor',
  'ItemBackgroundColor',
  'MinValue',
  'MaxValue',
  'ThumbPosition',
  'ThumbEnabled',
  'ColorLeft',
  'ColorRight',
  'ThumbColor',
  'NumberOfSteps',
  'ThumbColorActive',
  'ThumbColorInactive',
  'TrackColorActive',
  'TrackColorInactive',
  'Day',
  'Month',
  'MonthInText',
  'Year',
  'Hour',
  'Minute',
  'Instant',
  'InstantMillis',
  'SelectionColor',
  'SelectionDetailText',
  'SelectionImage',
  'TextColorDetail',
  'FontBoldDetail',
  'FontItalicDetail',
  'ListData',
  'ListViewLayout',
  'Orientation',
  'ImageWidth',
  'ImageHeight',
  'ElementColor',
  'DividerColor',
  'DividerThickness',
  'ElementMarginsWidth',
  'ElementCornerRadius',
  'BounceEdgeEffect',
  'MultiSelect',
  'NotifierLength',
]);

export const SIMULATION_ACCEPTED_PROPS = new Set([
  ...SIMULATION_VISUAL_PROPS,
  ...SIMULATION_BEHAVIOR_PROPS,
]);

export const EVENT_ARGS = {
  Slider: { PositionChanged: ['thumbPosition'] },
  Spinner: { AfterSelecting: ['selection'] },
  Canvas: {
    TouchDown: ['x', 'y'],
    TouchUp: ['x', 'y'],
    Touched: ['x', 'y', 'touchedAnySprite'],
  },
};

export function isSimulationSupportedType(type) {
  return SIMULATION_SUPPORTED_TYPES.has(type);
}

export function isSimulationNonVisibleType(type) {
  return SIMULATION_NONVISIBLE_TYPES.has(type);
}

export function isSimulationAcceptedProp(componentType, propName) {
  if (SIMULATION_IGNORED_PROPS_BY_TYPE[componentType]?.has(propName)) return false;
  return SIMULATION_ACCEPTED_PROPS.has(propName);
}

export function normalizeElements(value) {
  if (Array.isArray(value)) return value.map(item => normalizeElementItem(item));
  if (value && typeof value === 'object') {
    return Object.values(value).map(item => normalizeElementItem(item));
  }
  return elementsFromString(value);
}

export function elementsFromString(value) {
  const text = String(value ?? '').trim();
  if (!text) return [];
  return text.split(',').map(item => item.trim()).filter(Boolean);
}

const LISTVIEW_LAYOUT_COLUMNS = {
  0: ['Text1'],
  1: ['Text1', 'Text2'],
  2: ['Text1', 'Text2'],
  3: ['Text1', 'Image'],
  4: ['Text1', 'Text2', 'Image'],
  5: ['Text1', 'Text2', 'Image'],
};

const LISTVIEW_ROW_KEYS = new Set([
  'Text1',
  'Text2',
  'Image',
  'MainText',
  'DetailText',
  'ImageName',
]);

function listViewColumnsForLayout(layoutValue = 0) {
  return LISTVIEW_LAYOUT_COLUMNS[Number(layoutValue)] || LISTVIEW_LAYOUT_COLUMNS[0];
}

function rowDisplayText(row) {
  return String(row?.Text1 ?? row?.MainText ?? row?.text ?? row?.value ?? '');
}

function attachRowStringifier(row) {
  if (!row || typeof row !== 'object') return row;
  return Object.defineProperties(row, {
    toString: {
      value() { return rowDisplayText(this); },
      enumerable: false,
      configurable: true,
    },
    valueOf: {
      value() { return rowDisplayText(this); },
      enumerable: false,
      configurable: true,
    },
    [Symbol.toPrimitive]: {
      value() { return rowDisplayText(this); },
      enumerable: false,
      configurable: true,
    },
  });
}

function isStructuredListViewRow(item) {
  return Boolean(item && typeof item === 'object' && Object.keys(item).some(key => LISTVIEW_ROW_KEYS.has(key)));
}

function normalizeListViewRow(item, layoutValue = 0) {
  const source = item && typeof item === 'object' ? item : { Text1: item };
  const row = {};
  for (const column of listViewColumnsForLayout(layoutValue)) {
    const alias = column === 'Text1' ? 'MainText' : column === 'Text2' ? 'DetailText' : 'ImageName';
    const fallback = column === 'Image' ? 'None' : '';
    row[column] = String(source[column] ?? source[alias] ?? fallback);
  }
  for (const [key, itemValue] of Object.entries(source)) {
    if (!(key in row)) row[key] = itemValue;
  }
  return attachRowStringifier(row);
}

function normalizeElementItem(item) {
  if (isStructuredListViewRow(item)) return normalizeListViewRow(item);
  return String(item ?? '');
}

export function parseListData(value, layoutValue = 0) {
  if (Array.isArray(value)) return value.map(row => normalizeListViewRow(row, layoutValue));
  const text = String(value ?? '').trim();
  if (!text) return [];
  let parsed;
  try {
    parsed = JSON.parse(text);
  } catch {
    return [];
  }
  if (!Array.isArray(parsed)) return [];
  return parsed.map(row => normalizeListViewRow(row, layoutValue));
}

function monthName(month) {
  return MONTH_NAMES[Number(month) - 1] || '';
}

function isBooleanProp(propName) {
  return [
    'Visible',
    'Enabled',
    'Checked',
    'On',
    'NumbersOnly',
    'PasswordVisible',
    'ThumbEnabled',
    'Clickable',
    'MultiLine',
    'ReadOnly',
    'Scrollable',
    'TitleVisible',
    'ActionBar',
    'HighContrast',
    'BigDefaultText',
    'ShowStatusBar',
    'ShowFeedback',
    'ScalePictureToFit',
    'ShowFilterBar',
    'FontBold',
    'FontItalic',
    'FontBoldDetail',
    'FontItalicDetail',
    'HasMargins',
    'HTMLFormat',
    'BounceEdgeEffect',
    'MultiSelect',
  ].includes(propName);
}

function isNumericProp(propName) {
  return [
    'Width',
    'Height',
    'Left',
    'Top',
    'FontSize',
    'FontSizeDetail',
    'RotationAngle',
    'SelectionIndex',
    'MinValue',
    'MaxValue',
    'ThumbPosition',
    'NumberOfSteps',
    'Day',
    'Month',
    'Year',
    'Hour',
    'Minute',
    'InstantMillis',
    'AlignHorizontal',
    'AlignVertical',
    'TextAlignment',
    'Shape',
    'Scaling',
    'ListViewLayout',
    'Orientation',
    'ImageWidth',
    'ImageHeight',
    'DividerThickness',
    'ElementMarginsWidth',
    'ElementCornerRadius',
    'NotifierLength',
  ].includes(propName);
}

function coerceBoolean(value) {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number' && Number.isFinite(value)) return value !== 0;
  const text = String(value ?? '').trim().toLowerCase();
  if (['true', 't', 'yes', 'y', '1', 'on'].includes(text)) return true;
  if (['false', 'f', 'no', 'n', '0', 'off', ''].includes(text)) return false;
  return value;
}

function coerceNumber(value) {
  if (typeof value === 'number') return Number.isFinite(value) ? value : value;
  if (typeof value === 'boolean') return value ? 1 : 0;
  const text = String(value ?? '').trim();
  if (!text) return value;
  const numberValue = Number(text);
  return Number.isFinite(numberValue) ? numberValue : value;
}

export function coerceSimulationValue(componentType, propName, value) {
  if (value === null || value === undefined) return value;
  if (Array.isArray(value)) return propName === 'Elements' ? normalizeElements(value) : value;
  if (isBooleanProp(propName)) return coerceBoolean(value);
  if (propName === 'Elements') return normalizeElements(value);
  if (propName === 'ElementsFromString') return String(value ?? '');
  if (propName === 'ListData') return typeof value === 'string' ? value : JSON.stringify(value);
  if (propName === 'Instant' && typeof value === 'string') return value;
  if (isNumericProp(propName)) return coerceNumber(value);
  return value;
}

export function deriveStateFromDesignerProps(componentType, props = {}) {
  const next = {};
  for (const [key, value] of Object.entries(props)) {
    if (!isSimulationAcceptedProp(componentType, key)) continue;
    next[key] = coerceSimulationValue(componentType, key, value);
    if (key === 'ElementsFromString') next.Elements = elementsFromString(value);
    if (componentType === 'ListView' && key === 'ListData') {
      const rows = parseListData(value, props.ListViewLayout ?? next.ListViewLayout ?? SIMULATION_DEFAULTS.ListView.ListViewLayout);
      if (rows.length) next.Elements = rows;
    }
    if (componentType === 'DatePicker' && ['Year', 'Month', 'Day'].includes(key)) {
      const year = Number(key === 'Year' ? next.Year : next.Year ?? props.Year ?? SIMULATION_DEFAULTS.DatePicker.Year);
      const month = Number(key === 'Month' ? next.Month : next.Month ?? props.Month ?? SIMULATION_DEFAULTS.DatePicker.Month);
      const day = Number(key === 'Day' ? next.Day : next.Day ?? props.Day ?? SIMULATION_DEFAULTS.DatePicker.Day);
      if (Number.isFinite(year) && Number.isFinite(month) && Number.isFinite(day)) {
        next.MonthInText = monthName(month);
        next.Instant = dateInstant(year, month, day);
      }
    }
    if (componentType === 'TimePicker' && ['Hour', 'Minute'].includes(key)) {
      const hour = Number(key === 'Hour' ? next.Hour : next.Hour ?? props.Hour ?? SIMULATION_DEFAULTS.TimePicker.Hour);
      const minute = Number(key === 'Minute' ? next.Minute : next.Minute ?? props.Minute ?? SIMULATION_DEFAULTS.TimePicker.Minute);
      if (Number.isFinite(hour) && Number.isFinite(minute)) next.Instant = timeInstant(hour, minute);
    }
  }
  return next;
}

export function resolveAssetUrl(designAssets = [], assetName = '') {
  const wanted = String(assetName || '').trim();
  if (!wanted) return '';
  const asset = (designAssets || []).find(item => {
    const name = typeof item === 'string' ? item : item?.name;
    return name === wanted || name?.split('/').pop() === wanted;
  });
  return typeof asset === 'object' ? asset?.url || '' : '';
}
