# The files of a design

## The folder

`design/<slug>/` holds the artboards (`NN-<name>.html`, kebab-case), the screenshots
they were read against, `design.json`, and `comments.json` when the canvas collected
comments. A new brief is a new slug; a design that already exists is never overwritten.

## design.json

```json
{
  "title": "Pricing page",
  "brief": "Pricing page for a note-taking app aimed at students",
  "feature": "odd/tasks/pricing-page.md",
  "artboards": [
    { "file": "01-comparison-table.html", "title": "Comparison table", "tradeoff": "Dense and scannable; weaker emotional pull", "width": 1440, "height": 900, "screenshot": ".screenshots/01.png" },
    { "file": "02-single-plan-hero.html", "title": "Single plan hero", "tradeoff": "Strong focus; hides plan differences", "width": 1440, "height": 900, "screenshot": ".screenshots/02.png" }
  ]
}
```

Fields: `title`, `brief`, optional `feature`, and `artboards[]` with `file`, `title`,
`tradeoff`, `width`, `height`, `screenshot`. `width` and `height` are the artboard's
intended viewport in CSS pixels. Keep the manifest in sync with the files on every change.

## comments.json

```json
{
  "version": 1,
  "comments": [
    { "id": "k3x9p2ab", "file": "02-bold.html", "selector": "body > header > h1", "excerpt": "Ship faster", "comment": "Make the headline shorter.", "createdAt": "2026-10-26T10:00:00.000Z", "status": "open", "sentAt": null, "resolvedAt": null },
    { "id": "m7q2w9cd", "file": "02-bold.html", "selector": "#cta", "excerpt": "Start", "comment": "The button jumps during the hover transition.", "screenshot": ".screenshots/m7q2w9cd.png", "createdAt": "2026-10-26T10:05:00.000Z", "status": "open", "sentAt": null, "resolvedAt": null }
  ]
}
```

`screenshot` is optional: a path relative to the design folder, under `.screenshots/`,
present only when the user attached an image. Never delete or reorder entries, and keep
the file valid JSON: the canvas refuses to overwrite a file it cannot parse.

## The screenshots

Every artboard carries the screenshot it was read against, named in the manifest's
`screenshot`, taken after the last change. `odachi design list` names the ones without it.
