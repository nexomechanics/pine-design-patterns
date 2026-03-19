# Pine Design Patterns API

Generates pine script boilerplate based on selected preset, declaration type, mode, and sections. Can be used to quickly get set up and start coding in pine script.

Tradingview's own templates are not composable. This API is an adaptive solution to that, the templates returned can be customised to match what you want to build.

A feature that can improve this API is implementing types of presets. For example, instead of just a generic "Strategy", we can have a "Triple TP Strategy". If you regularly code in pine script, this API can be very helpful. You can also access the it through a ready-made frontend at https://tools.nexomechanics.com/pine-design-patterns


## Endpoints

### `GET /`

Health check.

**Response**
```json
{ "ok": true }
```

### `POST /generate`

Generate a Pine Script skeleton.

**Headers**
```
Content-Type: application/json
X-API-Key: <key>   # only required if API_KEY env var is set
```

**Body**

| Field         | Type       | Required | Description |
|---------------|------------|----------|-------------|
| `preset`      | `string`   | No       | `"indicator"`, `"strategy"`, `"library"`, or `"custom"`. When not `"custom"`, sections and declaration are resolved automatically. |
| `declaration` | `string`   | No*      | `"indicator"`, `"strategy"`, or `"library"`. Required when `preset` is `"custom"`. Defaults to `"indicator"`. |
| `mode`        | `string`   | No       | `"prefilled"` (default) or `"minimal"`. Prefilled includes example code; minimal outputs empty section blocks only. |
| `sections`    | `string[]` | No*      | Ordered list of section keys. Only used when `preset` is `"custom"`. |

**Section keys**

`version`, `declaration`, `imports`, `constants`, `inputs`, `types`, `functions`, `logic`, `strategy_calls`, `plots`, `alerts`

**Response**
```json
{ "result": "// generated Pine Script..." }
```

## Examples

### Preset — Indicator
```bash
curl -X POST https://<domain>/generate \
  -H "Content-Type: application/json" \
  -d '{"preset": "indicator", "mode": "prefilled"}'
```

### Preset — Strategy (minimal)
```bash
curl -X POST https://<domain>/generate \
  -H "Content-Type: application/json" \
  -d '{"preset": "strategy", "mode": "minimal"}'
```

### Custom sections
```bash
curl -X POST https://<domain>/generate \
  -H "Content-Type: application/json" \
  -d '{
    "preset": "custom",
    "declaration": "indicator",
    "mode": "prefilled",
    "sections": ["version", "declaration", "inputs", "logic", "plots"]
  }'
```

### With API key
```bash
curl -X POST https://<domain>/generate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{"preset": "indicator"}'
```

## Context-aware generation

The API automatically adapts generated code based on which sections are active:

- **`constants` present** → `plots` uses `BULL_COLOR` / `BEAR_COLOR`; without it, uses `color.green` / `color.red`
- **`inputs` present** → `logic` uses `i_source` / `i_length`; function signature drops defaults; called with `calcMA(i_source, i_length)`
- **`inputs` absent** → function signature keeps defaults `(src = close, len = 14)`; called with `calcMA()`
- **`functions` absent** → `logic` uses `ta.sma(...)` directly
- **`logic` absent** → `plots` / `alerts` / `strategy_calls` inline all expressions
- **`library` declaration** → function signatures include `float` / `int` types and `export` prefix
