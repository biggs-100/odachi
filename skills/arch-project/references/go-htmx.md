# A Go project with server-rendered HTML (HTMX)

The shared shape of `references/layout.md` does not change: the domain stays pure,
the router is an adapter, and the fragments are templates. Three choices are worth
making once, before the first page.

## HTMX: a stable version, pinned

- Pin the newest stable htmx. A beta line changed event shapes and attribute
  inheritance and broke fragments that were working, so a beta is not a version to
  adopt in a project that has no time to debug somebody else's release.
- After any upgrade, re-verify the event shapes (`ev.detail`), the inheritance and
  the fragment swap before the release, not after it.
- The server returns the fragment alone: a wrapper that repeats the container's id
  is how one swap becomes two.

## Templates: templ, when the page has more than a couple of files

- Write the templates as `.templ` components compiled to typed Go (`templ generate`).
  A missing field is a compile error instead of a page that fails when a user opens
  it, which is the cheapest test a page can have.
- Components take explicit arguments and render a fragment directly from the
  adapter, so a receipt, a KPI or a table is one component and not a string.
- `html/template` is fine under three templates. Above that, the compiler earns its
  keep.

## The router lives in the adapter

- Prefer `net/http` or chi. gin, fiber and echo invite a business rule into a
  handler, and a rule in a handler is a rule the domain cannot test.
- The domain never imports the router: the guard of `references/guard-recipes.md`
  is what makes that true instead of intended.

## Verifying a fragment

The suite covers the handler and the component; it does not see the page. A
fragment is verified by the layer `references/ui-verification.md` describes - one
command, a real browser, an assertion - and after an htmx upgrade that walk is
repeated before the release.
