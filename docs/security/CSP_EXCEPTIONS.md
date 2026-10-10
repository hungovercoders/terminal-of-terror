# CSP exceptions

The site's Content-Security-Policy is set once, for every page (`/*`), in `public/_headers`. This file registers any
page that needs a **relaxed** policy on top of that baseline. `hygiene:csp-exceptions` fails if a per-path CSP in
`public/_headers` is missing here, or if an entry here no longer matches one there.

## Baseline (`/*`)

No inline styles or scripts, nothing framed, and only these third-party origins:

- `https://cdnjs.cloudflare.com`: the highlight.js script and stylesheet on Night School lessons, both pinned with
  SRI hashes in `site/templates/lesson.html`.
- `https://api.github.com`: `site.js` reads the latest release to point the download buttons at its files. No data
  leaves the site beyond the request itself.

## Exceptions

None. To add one, put a `### /path` heading here with these fields, then the matching rule in `public/_headers`:
Origin allowed, Directives added, Loader SRI, Why, Approved by, Data leaving site, Refresh policy.
