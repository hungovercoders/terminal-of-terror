// Renders docs/demos/social-card.html, the link-preview image for the website,
// to docs/assets/social-card.png at 1200×630 (the size Open Graph and Twitter
// cards expect), then shrinks it with pngquant when that is on PATH: some
// link-preview crawlers skip images over about 300 KB, and the full-colour
// render is three times that.
//
//   npm install            # once, for @playwright/test
//   brew install pngquant  # or apt install pngquant, for the small file
//   apt install fonts-noto-color-emoji  # the only emoji font the card may use
//   node docs/demos/social-card.mjs
//
// Chromium comes from Playwright's own install (npx playwright install chromium)
// or from PLAYWRIGHT_CHROMIUM, a path to a chromium binary.
import { chromium } from "@playwright/test";
import { fileURLToPath, pathToFileURL } from "node:url";
import { spawnSync } from "node:child_process";
import path from "node:path";

const demos = path.dirname(fileURLToPath(import.meta.url));
const source = path.join(demos, "social-card.html");
const out = path.join(demos, "..", "assets", "social-card.png");

const browser = await chromium.launch({ executablePath: process.env.PLAYWRIGHT_CHROMIUM || undefined });
const page = await browser.newPage({ viewport: { width: 1200, height: 630 }, deviceScaleFactor: 1 });
await page.goto(pathToFileURL(source).href);
await page.evaluate(() => document.fonts.ready);
// Without Noto the browser would quietly draw the system's emoji instead, and
// Apple's and Microsoft's may not be redistributed in an image. Ask Chromium
// which fonts it actually drew every element with, so an emoji Noto lacks, or
// one outside the emoji elements, is caught too.
const noto = await page.evaluate(async () => {
  await document.fonts.load('42px "Card Emoji"', "🎃").catch(() => {});
  return [...document.fonts].some((f) => f.family.replaceAll('"', "") === "Card Emoji" && f.status === "loaded");
});
const cdp = await page.context().newCDPSession(page);
await cdp.send("DOM.enable");
await cdp.send("CSS.enable");
const { root } = await cdp.send("DOM.getDocument");
const { nodeIds } = await cdp.send("DOM.querySelectorAll", { nodeId: root.nodeId, selector: "body *" });
const drawn = new Set();
for (const nodeId of nodeIds) {
  const { fonts } = await cdp.send("CSS.getPlatformFontsForNode", { nodeId });
  for (const font of fonts) drawn.add(font.familyName);
}
const others = [...drawn].filter((name) => /emoji/i.test(name) && name !== "Noto Color Emoji");
if (!noto || others.length > 0) {
  await browser.close();
  console.error(noto
    ? `The card draws emoji with ${others.join(", ")}: use only emoji Noto Color Emoji has, in elements set to var(--emoji).`
    : "Noto Color Emoji isn't installed, so the card can't be rendered: install it (apt install fonts-noto-color-emoji) and run this again.");
  process.exit(1);
}
await page.screenshot({ path: out, type: "png", clip: { x: 0, y: 0, width: 1200, height: 630 } });
await browser.close();

const shrink = spawnSync("pngquant", ["--force", "--quality", "80-100", "256", "--output", out, out], { stdio: "inherit" });
if (shrink.error) {
  console.warn("pngquant not found: the card is full size. Install it and run this again.");
}
console.log(`→ ${path.relative(process.cwd(), out)}`);
