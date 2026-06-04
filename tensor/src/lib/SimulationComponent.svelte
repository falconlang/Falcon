<script>
  import { createEventDispatcher } from 'svelte';
  import './simulation/components.css';
  import { isSimulationNonVisibleType } from './simulation-capabilities.js';
  import {
    isSimulationEnabled,
    isSimulationVisible,
  } from './design-schema-tree.js';
  import { forwardSvelteEvent } from './simulation/events.js';

  import AbsoluteArrangement from './simulation/components/AbsoluteArrangement.svelte';
  import Ball from './simulation/components/Ball.svelte';
  import Button from './simulation/components/Button.svelte';
  import Canvas from './simulation/components/Canvas.svelte';
  import Chart from './simulation/components/Chart.svelte';
  import ChartData2D from './simulation/components/ChartData2D.svelte';
  import CheckBox from './simulation/components/CheckBox.svelte';
  import Circle from './simulation/components/Circle.svelte';
  import CircularProgress from './simulation/components/CircularProgress.svelte';
  import ContactPicker from './simulation/components/ContactPicker.svelte';
  import DatePicker from './simulation/components/DatePicker.svelte';
  import EmailPicker from './simulation/components/EmailPicker.svelte';
  import EmptyComponent from './simulation/components/EmptyComponent.svelte';
  import FeatureCollection from './simulation/components/FeatureCollection.svelte';
  import FilePicker from './simulation/components/FilePicker.svelte';
  import Form from './simulation/components/Form.svelte';
  import HorizontalArrangement from './simulation/components/HorizontalArrangement.svelte';
  import HorizontalScrollArrangement from './simulation/components/HorizontalScrollArrangement.svelte';
  import Image from './simulation/components/Image.svelte';
  import ImagePicker from './simulation/components/ImagePicker.svelte';
  import ImageSprite from './simulation/components/ImageSprite.svelte';
  import Label from './simulation/components/Label.svelte';
  import LinearProgress from './simulation/components/LinearProgress.svelte';
  import LineString from './simulation/components/LineString.svelte';
  import ListPicker from './simulation/components/ListPicker.svelte';
  import ListView from './simulation/components/ListView.svelte';
  import Map from './simulation/components/Map.svelte';
  import Marker from './simulation/components/Marker.svelte';
  import PasswordTextBox from './simulation/components/PasswordTextBox.svelte';
  import PhoneNumberPicker from './simulation/components/PhoneNumberPicker.svelte';
  import Polygon from './simulation/components/Polygon.svelte';
  import Rectangle from './simulation/components/Rectangle.svelte';
  import Screen from './simulation/components/Screen.svelte';
  import Slider from './simulation/components/Slider.svelte';
  import Spinner from './simulation/components/Spinner.svelte';
  import Switch from './simulation/components/Switch.svelte';
  import TableArrangement from './simulation/components/TableArrangement.svelte';
  import TextBox from './simulation/components/TextBox.svelte';
  import TimePicker from './simulation/components/TimePicker.svelte';
  import Trendline from './simulation/components/Trendline.svelte';
  import UnsupportedComponent from './simulation/components/UnsupportedComponent.svelte';
  import VerticalArrangement from './simulation/components/VerticalArrangement.svelte';
  import VerticalScrollArrangement from './simulation/components/VerticalScrollArrangement.svelte';
  import VideoPlayer from './simulation/components/VideoPlayer.svelte';
  import WebViewer from './simulation/components/WebViewer.svelte';

  export let node;
  export let state = {};
  export let unsupported = [];
  export let assets = [];
  export let parentType = '';
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  const COMPONENTS = {
    AbsoluteArrangement,
    Ball,
    Button,
    Canvas,
    Chart,
    ChartData2D,
    CheckBox,
    Circle,
    CircularProgress,
    ContactPicker,
    DatePicker,
    EmailPicker,
    FeatureCollection,
    FilePicker,
    Form,
    HorizontalArrangement,
    HorizontalScrollArrangement,
    Image,
    ImagePicker,
    ImageSprite,
    Label,
    LinearProgress,
    LineString,
    ListPicker,
    ListView,
    Map,
    Marker,
    PasswordTextBox,
    PhoneNumberPicker,
    Polygon,
    Rectangle,
    Screen,
    Slider,
    Spinner,
    Switch,
    TableArrangement,
    TextBox,
    TimePicker,
    Trendline,
    VerticalArrangement,
    VerticalScrollArrangement,
    VideoPlayer,
    WebViewer,
  };

  $: props = state?.[node?.name] ?? {};
  $: visible = isSimulationVisible(state, node?.name);
  $: enabled = isSimulationEnabled(state, node?.name);
  $: nonVisible = isSimulationNonVisibleType(node?.type);
  $: unsupportedHere = unsupported.some(entry => entry.detail === `${node?.type}.${node?.name}`);
  $: Component = COMPONENTS[node?.type] || UnsupportedComponent;
  $: componentProps = propsForComponent(node?.type);
  $: slotChildren = childrenForSlot(node);

  function propsForComponent(type) {
    const base = { node, props, assets, parentType, unsupportedHere };
    if (!COMPONENTS[type]) return { node, props, assets, parentType };

    switch (type) {
      case 'Ball':
      case 'ImageSprite':
      case 'FeatureCollection':
      case 'ChartData2D':
      case 'Trendline':
        return { node };
      case 'Marker':
      case 'Circle':
      case 'LineString':
      case 'Polygon':
      case 'Rectangle':
        return { node, props, assets, parentType };
      case 'Canvas':
        return { ...base, state, enabled, actions, eventRunner };
      case 'Chart':
        return { ...base, state, eventRunner };
      case 'Map':
        return { ...base, state, actions, eventRunner };
      case 'Button':
      case 'Spinner':
      case 'ListPicker':
      case 'DatePicker':
      case 'TimePicker':
      case 'ImagePicker':
      case 'FilePicker':
      case 'ContactPicker':
      case 'PhoneNumberPicker':
        return { ...base, enabled, visible, actions, eventRunner };
      case 'TextBox':
      case 'PasswordTextBox':
      case 'EmailPicker':
        return { ...base, enabled, actions, eventRunner };
      case 'CheckBox':
      case 'Switch':
      case 'Slider':
        return { ...base, enabled, eventRunner };
      case 'Image':
        return { ...base, visible, eventRunner };
      case 'WebViewer':
      case 'VideoPlayer':
        return { ...base, actions, eventRunner };
      default:
        return base;
    }
  }

  function childrenForSlot(current) {
    const children = current?.children || [];
    if (current?.type === 'Canvas') return children.filter(child => child.type === 'Ball' || child.type === 'ImageSprite');
    if ([
      'Screen',
      'Form',
      'VerticalArrangement',
      'VerticalScrollArrangement',
      'HorizontalArrangement',
      'HorizontalScrollArrangement',
      'AbsoluteArrangement',
      'TableArrangement',
    ].includes(current?.type)) {
      return children;
    }
    return [];
  }

  function childEvent(event) {
    forwardSvelteEvent(dispatch, event);
  }
</script>

{#if visible && !nonVisible}
  <svelte:component
    this={Component || EmptyComponent}
    {...componentProps}
    on:event={childEvent}
    on:property={childEvent}
    on:interaction={childEvent}
  >
    {#each slotChildren as child (child.pathId || child.name)}
      <svelte:self
        node={child}
        {state}
        {unsupported}
        {assets}
        {actions}
        {eventRunner}
        parentType={node.type}
        on:event={childEvent}
        on:property={childEvent}
        on:interaction={childEvent}
      />
    {/each}
  </svelte:component>
{/if}
