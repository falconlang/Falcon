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
  // Newly supported components
  'CircularProgress',
  'LinearProgress',
  'TableArrangement',
  'WebViewer',
  'VideoPlayer',
  'EmailPicker',
  'ImagePicker',
  'FilePicker',
  'ContactPicker',
  'PhoneNumberPicker',
  'Canvas',
  'Ball',
  'ImageSprite',
  'Chart',
  'ChartData2D',
  'Trendline',
  'Map',
  'Marker',
  'Circle',
  'LineString',
  'Polygon',
  'Rectangle',
  'FeatureCollection',
]);

export const SIMULATION_NONVISIBLE_TYPES = new Set([
  'Notifier',
  'TinyDB',
  'ChartData2D',
  'Trendline',
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
    CircularProgress: {
      Visible: true,
      Width: -1,
      Height: -1,
      Color: '&HFF0000FF',
    },
    LinearProgress: {
      Visible: true,
      Width: -2,
      Height: -1,
      ProgressColor: '&HFF0000FF',
      IndeterminateColor: '&HFF0000FF',
      Indeterminate: true,
      Minimum: 0,
      Maximum: 100,
      Progress: 0,
    },
    TableArrangement: {
      Visible: true,
      Columns: 2,
      Rows: 2,
      Width: -2,
      Height: -1,
      BackgroundColor: 'transparent',
    },
    WebViewer: {
      Visible: true,
      Width: -2,
      Height: -2,
      HomeUrl: '',
      CurrentUrl: '',
      CurrentPageTitle: '',
      FollowLinks: true,
      IgnoreSslErrors: false,
      PromptforPermission: true,
    },
    VideoPlayer: {
      Visible: true,
      Width: -2,
      Height: -1,
      Source: '',
      Volume: 50,
      FullScreen: false,
      Loop: false,
    },
    EmailPicker: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      Hint: '',
      HintColor: '&HFF888888',
      BackgroundColor: '&HFFFFFFFF',
      Width: 160,
    },
    ImagePicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
      Selection: '',
    },
    FilePicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
      Selection: '',
      Action: 'Pick Existing File',
      MimeType: '*/*',
    },
    ContactPicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
      ContactName: '',
      ContactUri: '',
      EmailAddress: '',
      EmailAddressList: [],
      PhoneNumber: '',
      PhoneNumberList: [],
      PhoneNumberType: [],
      Picture: '',
    },
    PhoneNumberPicker: {
      ...COMMON_VISIBLE_PROPS,
      ...BUTTON_STYLE_PROPS,
      Text: '',
      PhoneNumber: '',
      ContactName: '',
      ContactUri: '',
      Picture: '',
      EmailAddress: '',
      EmailAddressList: [],
      PhoneNumberList: [],
      PhoneNumberType: [],
    },
    Canvas: {
      Visible: true,
      Width: -2,
      Height: 300,
      BackgroundColor: '&HFFFFFFFF',
      BackgroundImage: '',
      BackgroundImageinBase64: '',
      PaintColor: '&HFF000000',
      LineWidth: 2,
      FontSize: 14,
      TextAlignment: 1,
      TapThreshold: 15,
      ExtendMovesOutsideCanvas: false,
    },
    Ball: {
      Visible: true,
      X: 0,
      Y: 0,
      Z: 1,
      Radius: 5,
      PaintColor: '&HFF000000',
      Speed: 0,
      Heading: 0,
      Interval: 100,
      Enabled: true,
      OriginAtCenter: false,
    },
    ImageSprite: {
      Visible: true,
      X: 0,
      Y: 0,
      Z: 1,
      Width: -1,
      Height: -1,
      Picture: '',
      Speed: 0,
      Heading: 0,
      Interval: 100,
      Enabled: true,
      Rotates: true,
      OriginX: 0,
      OriginY: 0,
      MarkOrigin: '',
    },
    Chart: {
      Visible: true,
      Width: -2,
      Height: 300,
      Type: 0,
      BackgroundColor: '&HFFFFFFFF',
      AxesTextColor: '&HFF000000',
      GridEnabled: true,
      LegendEnabled: true,
      Description: '',
      Labels: [],
      PieRadius: 100,
      XFromZero: false,
      YFromZero: false,
    },
    ChartData2D: {
      Visible: false,
      Label: '',
      Color: '&HFF000000',
      LineType: 0,
      PointShape: 0,
      Elements: [],
      ElementsFromPairs: '',
      Colors: [],
      DataLabelColor: '&HFF000000',
      HighlightColor: '',
    },
    Trendline: {
      Visible: true,
      Color: '&HFF000000',
      Model: 'Linear',
      StrokeWidth: 1,
      StrokeStyle: 1,
      Extend: true,
      ChartData: '',
    },
    Map: {
      Visible: true,
      Width: -2,
      Height: 300,
      Latitude: 42.359144,
      Longitude: -71.093612,
      ZoomLevel: 13,
      MapType: 1,
      EnablePan: true,
      EnableZoom: true,
      EnableRotation: false,
      ShowCompass: false,
      ShowScale: false,
      ShowUser: false,
      ShowZoom: false,
      Rotation: 0,
      CustomUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
      BoundingBox: '',
      CenterFromString: '',
    },
    Marker: {
      Visible: true,
      Latitude: 0,
      Longitude: 0,
      Title: '',
      Description: '',
      FillColor: '&HFFFF0000',
      FillOpacity: 1,
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1,
      StrokeWidth: 1,
      Draggable: false,
      EnableInfobox: false,
      ImageAsset: '',
      AnchorHorizontal: 3,
      AnchorVertical: 3,
    },
    Circle: {
      Visible: true,
      Latitude: 0,
      Longitude: 0,
      Radius: 0,
      Title: '',
      Description: '',
      FillColor: '&HFFFF0000',
      FillOpacity: 1,
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1,
      StrokeWidth: 1,
      Draggable: false,
      EnableInfobox: false,
    },
    LineString: {
      Visible: true,
      Points: [],
      PointsFromString: '',
      Title: '',
      Description: '',
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1,
      StrokeWidth: 3,
      Draggable: false,
      EnableInfobox: false,
    },
    Polygon: {
      Visible: true,
      Points: [],
      PointsFromString: '',
      HolePoints: [],
      HolePointsFromString: '',
      Title: '',
      Description: '',
      FillColor: '&HFFFF0000',
      FillOpacity: 1,
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1,
      StrokeWidth: 1,
      Draggable: false,
      EnableInfobox: false,
    },
    Rectangle: {
      Visible: true,
      NorthLatitude: 0,
      SouthLatitude: 0,
      EastLongitude: 0,
      WestLongitude: 0,
      Title: '',
      Description: '',
      FillColor: '&HFFFF0000',
      FillOpacity: 1,
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1,
      StrokeWidth: 1,
      Draggable: false,
      EnableInfobox: false,
    },
    FeatureCollection: {
      Visible: true,
      Width: -2,
      Height: -2,
      FeaturesFromGeoJSON: '',
      Source: '',
    },
  };
}

export const SIMULATION_DEFAULTS = buildSimulationDefaults();

export const SIMULATION_BEHAVIOR_PROPS = new Set([
  'Namespace',
  'DataSourceKey',
  'Source',
  'LocationSensor',
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
  'WidthPercent',
  'HeightPercent',
  'Left',
  'Top',
  'BackgroundColor',
  'BackgroundImage',
  'TextColor',
  'HintText',
  'FontSize',
  'TextSize',
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
  // CircularProgress / LinearProgress
  'Color',
  'ProgressColor',
  'IndeterminateColor',
  'Indeterminate',
  'Minimum',
  'Maximum',
  'Progress',
  // TableArrangement
  'Columns',
  'Rows',
  // WebViewer
  'HomeUrl',
  'CurrentUrl',
  'CurrentPageTitle',
  'FollowLinks',
  'IgnoreSslErrors',
  'PromptforPermission',
  // VideoPlayer
  'Source',
  'Volume',
  'FullScreen',
  'Loop',
  // Pickers
  'Selection',
  'ContactName',
  'ContactUri',
  'EmailAddress',
  'EmailAddressList',
  'PhoneNumber',
  'PhoneNumberList',
  'PhoneNumberType',
  'Picture',
  'Action',
  'MimeType',
  // Canvas
  'PaintColor',
  'LineWidth',
  'TapThreshold',
  'ExtendMovesOutsideCanvas',
  'BackgroundImageinBase64',
  // Ball / ImageSprite
  'X',
  'Y',
  'Z',
  'Radius',
  'Speed',
  'Heading',
  'Interval',
  'OriginAtCenter',
  'OriginX',
  'OriginY',
  'MarkOrigin',
  'Rotates',
  // Chart
  'Type',
  'AxesTextColor',
  'GridEnabled',
  'LegendEnabled',
  'Description',
  'Labels',
  'PieRadius',
  'XFromZero',
  'YFromZero',
  'Label',
  'LineType',
  'PointShape',
  'DataLabelColor',
  'Colors',
  'HighlightColor',
  'ElementsFromPairs',
  'ChartData',
  'CorrelationCoefficient',
  'LinearCoefficient',
  'RSquared',
  'YIntercept',
  'XIntercepts',
  'Predictions',
  'Results',
  'ExponentialBase',
  'ExponentialCoefficient',
  'LogarithmCoefficient',
  'LogarithmConstant',
  'QuadraticCoefficient',
  'Model',
  'StrokeStyle',
  'StrokeWidth',
  'Extend',
  // Map
  'Latitude',
  'Longitude',
  'ZoomLevel',
  'MapType',
  'EnablePan',
  'EnableZoom',
  'EnableRotation',
  'ShowCompass',
  'ShowScale',
  'ShowUser',
  'ShowZoom',
  'Rotation',
  'CustomUrl',
  'BoundingBox',
  'CenterFromString',
  // Map features
  'Title',
  'FillColor',
  'FillOpacity',
  'StrokeColor',
  'StrokeOpacity',
  'Draggable',
  'EnableInfobox',
  'ImageAsset',
  'AnchorHorizontal',
  'AnchorVertical',
  'Points',
  'PointsFromString',
  'HolePoints',
  'HolePointsFromString',
  'NorthLatitude',
  'SouthLatitude',
  'EastLongitude',
  'WestLongitude',
  'FeaturesFromGeoJSON',
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
    Dragged: ['startX', 'startY', 'prevX', 'prevY', 'currentX', 'currentY', 'draggedAnySprite'],
    Flung: ['x', 'y', 'speed', 'heading', 'xvel', 'yvel', 'flungSprite'],
  },
  Ball: {
    Touched: ['x', 'y'],
    TouchDown: ['x', 'y'],
    TouchUp: ['x', 'y'],
    Dragged: ['startX', 'startY', 'prevX', 'prevY', 'currentX', 'currentY'],
    Flung: ['x', 'y', 'speed', 'heading', 'xvel', 'yvel'],
    CollidedWith: ['other'],
    NoLongerCollidingWith: ['other'],
    EdgeReached: ['edge'],
  },
  ImageSprite: {
    Touched: ['x', 'y'],
    TouchDown: ['x', 'y'],
    TouchUp: ['x', 'y'],
    Dragged: ['startX', 'startY', 'prevX', 'prevY', 'currentX', 'currentY'],
    Flung: ['x', 'y', 'speed', 'heading', 'xvel', 'yvel'],
    CollidedWith: ['other'],
    NoLongerCollidingWith: ['other'],
    EdgeReached: ['edge'],
  },
  LinearProgress: { ProgressChanged: ['progress'] },
  Chart: { EntryClick: ['series', 'x', 'y'] },
  ChartData2D: { EntryClick: ['x', 'y'] },
  Trendline: { Updated: ['results'] },
  Map: {
    TapAtPoint: ['latitude', 'longitude'],
    LongPressAtPoint: ['latitude', 'longitude'],
    DoubleTapAtPoint: ['latitude', 'longitude'],
    FeatureClick: ['feature'],
    FeatureLongClick: ['feature'],
    FeatureDrag: ['feature'],
    FeatureStartDrag: ['feature'],
    FeatureStopDrag: ['feature'],
    GotFeatures: ['url', 'features'],
    LoadError: ['url', 'responseCode', 'errorMessage'],
    InvalidPoint: ['message'],
  },
  FeatureCollection: {
    FeatureClick: ['feature'],
    FeatureLongClick: ['feature'],
    FeatureDrag: ['feature'],
    FeatureStartDrag: ['feature'],
    FeatureStopDrag: ['feature'],
    GotFeatures: ['url', 'features'],
    LoadError: ['url', 'responseCode', 'errorMessage'],
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
    'Indeterminate',
    'FollowLinks',
    'IgnoreSslErrors',
    'PromptforPermission',
    'FullScreen',
    'Loop',
    'ExtendMovesOutsideCanvas',
    'OriginAtCenter',
    'Rotates',
    'GridEnabled',
    'LegendEnabled',
    'XFromZero',
    'YFromZero',
    'Extend',
    'EnablePan',
    'EnableZoom',
    'EnableRotation',
    'ShowCompass',
    'ShowScale',
    'ShowUser',
    'ShowZoom',
    'Draggable',
    'EnableInfobox',
  ].includes(propName);
}

function isNumericProp(propName) {
  return [
    'Width',
    'Height',
    'WidthPercent',
    'HeightPercent',
    'Left',
    'Top',
    'FontSize',
    'TextSize',
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
    'Minimum',
    'Maximum',
    'Progress',
    'Columns',
    'Rows',
    'Volume',
    'X',
    'Y',
    'Z',
    'Radius',
    'Speed',
    'Heading',
    'Interval',
    'OriginX',
    'OriginY',
    'LineWidth',
    'TapThreshold',
    'Type',
    'PieRadius',
    'LineType',
    'PointShape',
    'StrokeStyle',
    'StrokeWidth',
    'Latitude',
    'Longitude',
    'ZoomLevel',
    'MapType',
    'Rotation',
    'FillOpacity',
    'StrokeOpacity',
    'AnchorHorizontal',
    'AnchorVertical',
    'NorthLatitude',
    'SouthLatitude',
    'EastLongitude',
    'WestLongitude',
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

function parseOriginMarker(value) {
  const match = String(value ?? '').trim().match(/^\(\s*([-+]?\d+(?:\.\d+)?)\s*,\s*([-+]?\d+(?:\.\d+)?)\s*\)$/);
  if (!match) return null;
  const x = Number(match[1]);
  const y = Number(match[2]);
  if (!Number.isFinite(x) || !Number.isFinite(y)) return null;
  return [Math.max(0, Math.min(1, x)), Math.max(0, Math.min(1, y))];
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
    if (componentType === 'ImageSprite' && key === 'MarkOrigin') {
      const origin = parseOriginMarker(value);
      if (origin) {
        next.OriginX = origin[0];
        next.OriginY = origin[1];
      }
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
