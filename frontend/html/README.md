# koka-html

Type-safe HTML construction library for Koka.

Exposes an effect-based DSL that mirrors HTML structure:

```koka
import html/html

fun my-component()
  html
    head
      title
        text("Hello")
    body
      span(classes=["red"])
        text("World")
```

Renders:

```html
<html>
  <head>
    <title>
      Hello
    </title>
  </head>
  <body>
    <span class="red">
      World
    </span>
  </body>
</html>
```

## Modules

| Module           | Description                               |
|------------------|-------------------------------------------|
| `html-builder`   | Core `component` type, builder effect, render to string |
| `core-components`| Standard HTML element wrappers (`div`, `span`, `a`, `table`, etc.) |
| `escape`         | HTML entity escaping (`html-escape`)      |
| `helpers`        | Convenience: `text()`, `component()` wrappers |
| `layout`         | Layout components: `head`, `body`, `table`, `form` |
| `style-builder`  | Inline style DSL builder                  |

## Usage

```koka
import html/html/html    // re-exports html-builder + core-components + helpers
import html/html/escape  // html-escape
```

See `examples/` for standalone usage.

## License

MIT — Koka-Community Authors. See `LICENSE`.
