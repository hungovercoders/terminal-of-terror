/* Terminal of Terror · Channel 13 · site behaviour.
   Everything here is a garnish: the pages read fine with JavaScript off. */
(function () {
  "use strict";

  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var root = document.body.getAttribute("data-root") || "";
  var scrollMode = reduceMotion ? "auto" : "smooth";

  function $(sel, el) { return (el || document).querySelector(sel); }
  function $$(sel, el) { return Array.prototype.slice.call((el || document).querySelectorAll(sel)); }

  /* ---- tuning in: a burst of static the first time per visit ---- */
  function tuneIn() {
    var tuning = $("#tuning");
    if (!tuning || reduceMotion) return;
    var seen = false;
    try { seen = sessionStorage.getItem("tot-tuned") === "1"; } catch (e) { /* private mode */ }
    if (seen) return;
    tuning.classList.add("on");
    try { sessionStorage.setItem("tot-tuned", "1"); } catch (e) { /* ignore */ }
    setTimeout(function () {
      tuning.classList.add("off");
      setTimeout(function () { tuning.classList.remove("on", "off"); }, 500);
    }, 900);
  }

  /* ---- the hero: a still first, the 1.4 MB recording once the page is up ---- */
  function heroAnimation() {
    var hero = $("#hero");
    if (!hero || reduceMotion) return;
    window.addEventListener("load", function () {
      var gif = new Image();
      gif.onload = function () { hero.src = gif.src; };
      gif.src = hero.getAttribute("data-animated");
    });
  }

  /* ---- the VCR clock ---- */
  function vcrClock() {
    var clock = $("#clock");
    if (!clock) return;
    var tick = function () {
      var d = new Date();
      var h = d.getHours(), m = d.getMinutes();
      var colon = d.getSeconds() % 2 ? " " : ":";
      clock.textContent = (h < 10 ? "0" : "") + h + colon + (m < 10 ? "0" : "") + m;
    };
    tick();
    setInterval(tick, 1000);
  }

  /* ---- the host types his line ---- */
  function typewriter() {
    var tw = $(".typewriter");
    var line = tw && tw.getAttribute("data-type");
    if (!line || reduceMotion) return;
    tw.innerHTML = '📺 <em>"<span class="typed"></span>"</em><span class="caret"></span>';
    var typed = $(".typed", tw), i = 0;
    var step = function () {
      if (i <= line.length) {
        typed.textContent = line.slice(0, i++);
        setTimeout(step, line[i - 2] === "." || line[i - 2] === "," ? 180 : 28);
      } else {
        $(".caret", tw).remove();
      }
    };
    setTimeout(step, 1200);
  }

  /* ---- Night School: syntax highlighting ---- */
  // highlight.js is deferred after this script; every deferred script has
  // run by DOMContentLoaded.
  function highlight() {
    document.addEventListener("DOMContentLoaded", function () {
      if (window.hljs) window.hljs.highlightAll();
    });
  }

  /* ---- tabs ---- */
  function tabs() {
    $$("[data-tabs]").forEach(function (tabs) {
      var buttons = $$("[role=tab]", tabs), panels = $$("[data-panel]", tabs);
      buttons.forEach(function (b) {
        b.addEventListener("click", function () {
          buttons.forEach(function (x) { x.setAttribute("aria-selected", x === b ? "true" : "false"); });
          panels.forEach(function (p) { p.hidden = p.getAttribute("data-panel") !== b.getAttribute("data-tab"); });
        });
      });
    });
  }

  /* ---- copy buttons on terminal blocks ---- */
  function copyButtons() {
    if (!navigator.clipboard) return;
    $$("pre[data-copy]").forEach(function (pre) {
      var btn = document.createElement("button");
      btn.className = "copy";
      btn.type = "button";
      btn.textContent = "copy";
      btn.addEventListener("click", function () {
        var text = $$("code", pre).map(function (c) { return c.textContent; }).join("\n");
        // Drop the comments: nobody wants "# tune in" in their shell.
        text = text.split("\n").map(function (l) { return l.replace(/\s+#.*$/, ""); }).join("\n");
        navigator.clipboard.writeText(text).then(function () {
          btn.textContent = "copied!";
          setTimeout(function () { btn.textContent = "copy"; }, 1400);
        });
      });
      pre.appendChild(btn);
    });
  }

  /* ---- tonight: the countdown, the moon and the Fact of the Night ---- */
  function tonight() {
    var nights = $("#countdown-nights");
    if (!nights) return;
    var now = new Date();
    var today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    countdown(nights, today);
    moon(now);
    factOfTheNight(today);
  }

  function countdown(nights, today) {
    var hallow = new Date(today.getFullYear(), 9, 31);
    if (today > hallow) hallow = new Date(today.getFullYear() + 1, 9, 31);
    var n = Math.round((hallow - today) / 86400000);
    var label = $("#countdown-label");
    if (n === 0) {
      nights.textContent = "Tonight!";
      label.textContent = "Happy Halloween, fiends";
      return;
    }
    nights.textContent = String(n);
    label.textContent = n === 1 ? "night until Halloween" : "nights until Halloween";
    if (today.getMonth() === 9) label.textContent += " · Night " + today.getDate() + " of the 31 Nights of Fright";
  }

  // Mean lunar month from a known new moon, the same sum the CLI does.
  function moon(now) {
    var synodic = 29.530588853;
    var known = Date.UTC(2000, 0, 6, 18, 14);
    var age = ((now - known) / 86400000) % synodic;
    if (age < 0) age += synodic;
    var phases = [[0.5, "New Moon", "🌑"], [1.5, "Waxing Crescent", "🌒"], [2.5, "First Quarter", "🌓"], [3.5, "Waxing Gibbous", "🌔"], [4.5, "Full Moon", "🌕"], [5.5, "Waning Gibbous", "🌖"], [6.5, "Last Quarter", "🌗"], [7.5, "Waning Crescent", "🌘"], [8, "New Moon", "🌑"]];
    var eighths = age / synodic * 8;
    var phase = phases.filter(function (p) { return eighths < p[0]; })[0];
    if (phase) {
      $("#moon-name").textContent = phase[1];
      $("#moon-emoji").textContent = phase[2];
    }
    if (Math.abs(age - synodic / 2) < 1) $("#moon-tile").classList.add("full");
  }

  function factOfTheNight(today) {
    var factsEl = $("#facts");
    if (!factsEl) return;
    try {
      var facts = JSON.parse(factsEl.textContent);
      // A fact that stays put all day and changes tomorrow. It's picked
      // by this hash, not the CLI's seeded math/rand, so it need not be
      // the one `random --daily` shows; the tile says "Tonight's fact".
      var seed = today.getFullYear() * 10000 + (today.getMonth() + 1) * 100 + today.getDate();
      var x = seed;
      x = ((x >>> 16) ^ x) * 0x45d9f3b; x = ((x >>> 16) ^ x) * 0x45d9f3b; x = (x >>> 16) ^ x;
      var f = facts[Math.abs(x) % facts.length];
      $("#fact-text").textContent = f.emoji + " " + f.fact;
      var link = $("#fact-link");
      link.textContent = "More about " + f.name + " →";
      link.href = root + "monsters/" + f.id + ".html";
      $("#fact-date").textContent = today.toLocaleDateString(undefined, { weekday: "long", day: "numeric", month: "long" });
    } catch (e) { /* leave the placeholder */ }
  }

  /* ---- download links straight to the latest release's assets ---- */
  function downloads() {
    var downloads = $("#downloads");
    if (!downloads || !window.fetch) return;
    var repo = downloads.getAttribute("data-repo");
    // The list, not /releases/latest, which answers 404 until the first
    // release and leaves an error in the console.
    fetch("https://api.github.com/repos/" + repo + "/releases?per_page=10", { headers: { Accept: "application/vnd.github+json" } })
      .then(function (r) { return r.ok ? r.json() : []; })
      .then(function (list) {
        var rel = list.filter(function (r) { return !r.prerelease && !r.draft; })[0];
        if (!rel || !rel.assets) return;
        var v = $("#release-version");
        if (v) v.textContent = rel.tag_name;
        $$(".dl", downloads).forEach(function (a) {
          var key = a.getAttribute("data-asset");
          var asset = rel.assets.filter(function (x) { return x.name.indexOf("_" + key + ".") !== -1; })[0];
          if (asset) {
            a.href = asset.browser_download_url;
            a.title = asset.name;
          }
        });
        exampleVersion(rel.tag_name);
      })
      .catch(function () { /* the links already point at the releases page */ });
  }

  // Swap the example version in the text nodes only; the tag is data from
  // the API, never markup.
  function exampleVersion(tag) {
    var tar = $("#tar-example");
    if (!tar || !/^v?\d+\.\d+\.\d+$/.test(tag)) return;
    var ver = tag.replace(/^v/, "");
    var walker = document.createTreeWalker(tar, NodeFilter.SHOW_TEXT);
    while (walker.nextNode()) walker.currentNode.nodeValue = walker.currentNode.nodeValue.replace(/1\.0\.0/g, ver);
  }

  /* ---- monster pages ---- */
  function monsterPage() {
    var monster = $("article.monster");
    if (!monster) return;
    portraitFog();

    // Stat bars fill in.
    $$(".bar").forEach(function (b) { b.classList.add("pending"); });
    requestAnimationFrame(function () { setTimeout(function () { $$(".bar").forEach(function (b) { b.classList.remove("pending"); }); }, 150); });

    // Myth vs Movie: each card is a <details>, so clicking works without
    // us; r opens the lot, like the explorer.
    var myths = $$(".myth");
    var revealAll = function () { myths.forEach(function (m) { m.open = true; }); };
    var revealBtn = $("#reveal-all");
    if (revealBtn) revealBtn.addEventListener("click", revealAll);

    var tabs = $$(".section-tabs a");
    lightSectionTabs(tabs);
    explorerKeys(monster, tabs, revealAll);
  }

  // The portrait clears from the fog, like Guess the Monster.
  function portraitFog() {
    var portrait = $("#portrait");
    if (!portrait || reduceMotion) return;
    var art = portrait.getAttribute("data-art").split("\n");
    var width = Math.max.apply(null, art.map(function (l) { return l.length; }));
    var cells = [];
    art.forEach(function (l, y) { for (var x = 0; x < width; x++) cells.push([y, x]); });
    for (var i = cells.length - 1; i > 0; i--) { var j = Math.floor(Math.random() * (i + 1)); var t = cells[i]; cells[i] = cells[j]; cells[j] = t; }
    var fog = "░▒░ ░";
    var grid = art.map(function (l) { var row = []; for (var x = 0; x < width; x++) row.push(fog[(x * 7 + l.length) % fog.length]); return row; });
    var shown = 0, total = cells.length, start = null;
    var frame = function (ts) {
      if (!start) start = ts;
      var target = Math.min(total, Math.floor(total * Math.pow((ts - start) / 1100, 1.6)));
      while (shown < target) { var c = cells[shown++]; grid[c[0]][c[1]] = art[c[0]][c[1]] || " "; }
      portrait.textContent = grid.map(function (r) { return r.join(""); }).join("\n");
      if (shown < total) requestAnimationFrame(frame); else portrait.textContent = art.join("\n");
    };
    portrait.textContent = grid.map(function (r) { return r.join(""); }).join("\n");
    requestAnimationFrame(frame);
  }

  // Light the section tab for where you are.
  function lightSectionTabs(tabs) {
    if (!("IntersectionObserver" in window)) return;
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          tabs.forEach(function (a) { a.classList.toggle("active", a.getAttribute("href") === "#" + e.target.id); });
        }
      });
    }, { rootMargin: "-30% 0px -60% 0px" });
    tabs.forEach(function (a) { var s = $(a.getAttribute("href")); if (s) io.observe(s); });
  }

  // The explorer's keys, for people who never left the terminal.
  function explorerKeys(monster, tabs, revealAll) {
    var go = function (attr) { return function () { var to = monster.getAttribute(attr); if (to) location.href = to; }; };
    var scroll = function (top) { return function () { window.scrollBy({ top: top, behavior: scrollMode }); }; };
    var gallery = function () { location.href = "index.html"; };
    var actions = {
      h: go("data-prev"), ArrowLeft: go("data-prev"), p: go("data-prev"),
      l: go("data-next"), ArrowRight: go("data-next"), n: go("data-next"),
      g: gallery, Escape: gallery,
      r: revealAll,
      j: scroll(80), k: scroll(-80)
    };
    document.addEventListener("keydown", function (e) {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      var tag = (e.target.tagName || "").toLowerCase();
      if (tag === "input" || tag === "textarea") return;
      if (Object.prototype.hasOwnProperty.call(actions, e.key)) actions[e.key]();
      else if (e.key >= "1" && e.key <= "7") jumpToTab(tabs, e.key);
    });
  }

  // Tabs keep the explorer's numbers even when a monster has no film or
  // quotes, so find the one labelled with this key.
  function jumpToTab(tabs, key) {
    var a = tabs.filter(function (t) { var k = $("kbd", t); return k && k.textContent === key; })[0];
    if (a) $(a.getAttribute("href")).scrollIntoView({ behavior: scrollMode });
  }

  tuneIn();
  heroAnimation();
  vcrClock();
  typewriter();
  highlight();
  tabs();
  copyButtons();
  tonight();
  downloads();
  monsterPage();
})();
