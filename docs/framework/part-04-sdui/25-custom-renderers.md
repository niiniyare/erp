---
title: "Chapter 25: Custom Renderers"
part: "Part IV — The SDUI Layer"
chapter: 25
section: "25-custom-renderers"
related:
  - "[Chapter 21: SDUI Philosophy](21-sdui-philosophy.md)"
  - "[Chapter 22: Foundation Components](22-foundation-components.md)"
---

# Chapter 25: Custom Renderers

Custom renderers extend amis with new component types not available in the standard library. They are written in JavaScript/React and served as a separate JS bundle loaded alongside the amis SDK.

---

## 25.1. When a Custom Renderer Is Justified

### 25.1.1. Capability Gap

A custom renderer is justified when:
- A required UI pattern does not exist in amis's standard components
- An existing component can be configured to approximate the need, but the approximation is brittle or confusing to users

Examples of legitimate capability gaps:
- Real-time gauge display for fuel pump forecourt monitoring
- Interactive floor plan with clickable zones
- Gantt chart for project scheduling

### 25.1.2. Performance Gap

Some use cases require rendering performance that amis's generic components cannot achieve:
- A list of 10,000+ rows that must scroll smoothly (use a virtualised list renderer)
- A live-updating chart that updates every second (use a lightweight custom chart)

### 25.1.3. When to Push Back

Most "we need a custom renderer" requests can be satisfied with:
1. A different amis component (check the full amis docs before concluding something is impossible)
2. A custom `formatter` function on an existing column
3. A combination of existing components with conditional visibility

Custom renderers have a maintenance cost: they must be tested against each amis version upgrade, they require frontend expertise, and they cannot use amis's built-in form validation, API wiring, or permissions integration without reimplementing it.

---

## 25.2. Registering a Custom Renderer

### 25.2.1. The amis Custom Component API

Custom renderers use amis's `Renderer` decorator:

```jsx
// web/renderers/pump-status/index.jsx
import React from 'react';
import { Renderer } from 'amis';

@Renderer({ type: 'pump-status-display' })
class PumpStatusDisplay extends React.Component {
  render() {
    const { value, status, fuelType } = this.props.data;
    return (
      <div className={`pump-status pump-status--${status}`}>
        <span className="pump-id">{value}</span>
        <span className="fuel-type">{fuelType}</span>
        <span className={`status-indicator status--${status}`}>
          {status === 'active' ? '● Active' : '○ Idle'}
        </span>
      </div>
    );
  }
}
```

### 25.2.2. Serving the Custom Renderer JS Bundle

Custom renderer bundles are built with esbuild and served from `web/sdk/renderers/`:

```bash
# Build the renderer bundle
cd web/renderers/pump-status
esbuild index.jsx --bundle --outfile=../../sdk/renderers/pump-status.js \
  --external:react --external:amis --format=iife --global-name=PumpStatusRenderer
```

The bundle is loaded in `web/pages/index.html` after the amis SDK:

```html
<script src="/sdk/sdk.js"></script>
<script src="/sdk/renderers/pump-status.js"></script>
```

### 25.2.3. Registering the Renderer Type String

The type string (`pump-status-display`) is how the server-side page builder references the custom renderer in amis JSON:

```go
// In the page builder:
amis.Custom("pump-status-display").
    Data(map[string]interface{}{
        "value":    "${pump_id}",
        "status":   "${pump_status}",
        "fuelType": "${fuel_type}",
    })
```

This generates:
```json
{
  "type": "pump-status-display",
  "data": {
    "value": "${pump_id}",
    "status": "${pump_status}",
    "fuelType": "${fuel_type}"
  }
}
```

amis looks up the registered `Renderer` for type `pump-status-display` and invokes it.

---

## 25.3. Worked Examples

### 25.3.1. Forecourt Pump Status Renderer

For a petroleum retail tenant that needs real-time pump status display:

```jsx
@Renderer({ type: 'pump-status-grid' })
class PumpStatusGrid extends React.Component {
  state = { pumps: [], connected: false }

  componentDidMount() {
    // Connect to SSE stream for real-time updates
    this.eventSource = new EventSource('/api/v1/pumps/stream');
    this.eventSource.onmessage = (e) => {
      const pumps = JSON.parse(e.data);
      this.setState({ pumps, connected: true });
    };
  }

  componentWillUnmount() {
    this.eventSource?.close();
  }

  render() {
    const { pumps, connected } = this.state;
    if (!connected) return <div>Connecting to pump controller...</div>;

    return (
      <div className="pump-grid">
        {pumps.map(pump => (
          <PumpCard key={pump.id} pump={pump} />
        ))}
      </div>
    );
  }
}
```

The server-side page builder includes this in the dashboard:

```go
amis.Custom("pump-status-grid").
    Refresh(5000)  // fallback polling if SSE unavailable
```

### 25.3.2. Custom Chart Renderer with Apache ECharts

For complex chart types not available in amis's built-in chart component:

```jsx
import * as echarts from 'echarts';

@Renderer({ type: 'echarts-custom' })
class EChartsRenderer extends React.Component {
  chartRef = React.createRef();

  componentDidMount() {
    this.chart = echarts.init(this.chartRef.current, 'dark');
    this.updateChart();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.data !== this.props.data) {
      this.updateChart();
    }
  }

  updateChart() {
    const { option } = this.props.data;
    if (option) {
      this.chart.setOption(JSON.parse(option));
    }
  }

  render() {
    return <div ref={this.chartRef} style={{ width: '100%', height: 400 }} />;
  }
}
```

The server sends the full ECharts `option` object as a JSON string in the `option` field, giving the server-side page builder full control over the chart configuration without requiring a custom renderer change for each chart variant.

### 25.3.3. Map Renderer — Site Locations with Leaflet

```jsx
import L from 'leaflet';

@Renderer({ type: 'leaflet-map' })
class LeafletMap extends React.Component {
  mapRef = React.createRef();

  componentDidMount() {
    const { centerLat = -1.286389, centerLng = 36.817223, zoom = 10 } = this.props.data;

    this.map = L.map(this.mapRef.current).setView([centerLat, centerLng], zoom);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png').addTo(this.map);

    this.renderMarkers();
  }

  renderMarkers() {
    const { markers = [] } = this.props.data;
    markers.forEach(({ lat, lng, label, status }) => {
      const colour = status === 'active' ? 'green' : 'red';
      L.circleMarker([lat, lng], { color: colour, radius: 8 })
        .bindPopup(label)
        .addTo(this.map);
    });
  }

  render() {
    return <div ref={this.mapRef} style={{ width: '100%', height: 400 }} />;
  }
}
```

The page builder provides marker data from an API source:

```go
amis.Custom("leaflet-map").
    APISource("GET /api/v1/sites/map-data").
    DataMapping(map[string]string{
        "markers":   "$.data",
        "centerLat": "-1.286389",  // Nairobi
        "centerLng": "36.817223",
    })
```
