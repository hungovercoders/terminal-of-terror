/* Terminal of Terror · Channel 13 · site behaviour.
   Everything here is a garnish: the pages read fine with JavaScript off. */
(function () {
  "use strict";

  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var root = document.body.getAttribute("data-root") || "";

  function $(sel, el) { return (el || document).querySelector(sel); }
  function $$(sel, el) { return Array.prototype.slice.call((el || document).querySelectorAll(sel)); }

  /* ---- tuning in: a burst of static the first time per visit ---- */
  var tuning = $("#tuning");
  if (tuning && !reduceMotion) {
    var seen = false;
    try { seen = sessionStorage.getItem("tot-tuned") === "1"; } catch (e) { /* private mode */ }
    if (!seen) {
      tuning.classList.add("on");
      try { sessionStorage.setItem("tot-tuned", "1"); } catch (e) { /* ignore */ }
      setTimeout(function () {
        tuning.classList.add("off");
        setTimeout(function () { tuning.classList.remove("on", "off"); }, 500);
      }, 900);
    }
  }

  /* ---- the VCR clock ---- */
  var clock = $("#clock");
  if (clock) {
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
  var tw = $(".typewriter");
  if (tw && !reduceMotion) {
    var line = tw.getAttribute("data-type");
    if (line) {
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
  }

  /* ---- tabs ---- */
  $$("[data-tabs]").forEach(function (tabs) {
    var buttons = $$("[role=tab]", tabs), panels = $$("[data-panel]", tabs);
    buttons.forEach(function (b) {
      b.addEventListener("click", function () {
        buttons.forEach(function (x) { x.setAttribute("aria-selected", x === b ? "true" : "false"); });
        panels.forEach(function (p) { p.hidden = p.getAttribute("data-panel") !== b.getAttribute("data-tab"); });
      });
    });
  });

  /* ---- copy buttons on terminal blocks ---- */
  if (navigator.clipboard) {
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
  var nights = $("#countdown-nights");
  if (nights) {
    var now = new Date();
    var today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    var hallow = new Date(today.getFullYear(), 9, 31);
    if (today > hallow) hallow = new Date(today.getFullYear() + 1, 9, 31);
    var n = Math.round((hallow - today) / 86400000);
    var label = $("#countdown-label");
    if (n === 0) {
      nights.textContent = "Tonight!";
      label.textContent = "Happy Halloween, fiends";
    } else {
      nights.textContent = String(n);
      label.textContent = n === 1 ? "night until Halloween" : "nights until Halloween";
      if (today.getMonth() === 9) label.textContent += " · Night " + today.getDate() + " of the 31 Nights of Fright";
    }

    // Mean lunar month from a known new moon, the same sum the CLI does.
    var synodic = 29.530588853;
    var known = Date.UTC(2000, 0, 6, 18, 14);
    var age = ((now - known) / 86400000) % synodic;
    if (age < 0) age += synodic;
    var phases = [[0.5, "New Moon", "🌑"], [1.5, "Waxing Crescent", "🌒"], [2.5, "First Quarter", "🌓"], [3.5, "Waxing Gibbous", "🌔"], [4.5, "Full Moon", "🌕"], [5.5, "Waning Gibbous", "🌖"], [6.5, "Last Quarter", "🌗"], [7.5, "Waning Crescent", "🌘"], [8, "New Moon", "🌑"]];
    var eighths = age / synodic * 8;
    for (var p = 0; p < phases.length; p++) {
      if (eighths < phases[p][0]) {
        $("#moon-name").textContent = phases[p][1];
        $("#moon-emoji").textContent = phases[p][2];
        break;
      }
    }
    if (Math.abs(age - synodic / 2) < 1) $("#moon-tile").classList.add("full");

    var factsEl = $("#facts");
    if (factsEl) {
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
  }

  /* ---- download links straight to the latest release's assets ---- */
  var downloads = $("#downloads");
  if (downloads && window.fetch) {
    var repo = downloads.getAttribute("data-repo");
    fetch("https://api.github.com/repos/" + repo + "/releases/latest", { headers: { Accept: "application/vnd.github+json" } })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (rel) {
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
        var tar = $("#tar-example");
        if (tar && /^v?\d+\.\d+\.\d+$/.test(rel.tag_name)) {
          // Swap the example version in the text nodes only; the tag is
          // data from the API, never markup.
          var ver = rel.tag_name.replace(/^v/, "");
          var walker = document.createTreeWalker(tar, NodeFilter.SHOW_TEXT);
          while (walker.nextNode()) walker.currentNode.nodeValue = walker.currentNode.nodeValue.replace(/1\.0\.0/g, ver);
        }
      })
      .catch(function () { /* the links already point at the releases page */ });
  }

  /* ---- monster pages ---- */
  var monster = $("article.monster");
  if (monster) {
    // The portrait clears from the fog, like Guess the Monster.
    var portrait = $("#portrait");
    if (portrait && !reduceMotion) {
      var art = portrait.getAttribute("data-art").split("\n");
      var width = Math.max.apply(null, art.map(function (l) { return l.length; }));
      var cells = [];
      art.forEach(function (l, y) { for (var x = 0; x < width; x++) cells.push([y, x]); });
      for (var i = cells.length - 1; i > 0; i--) { var j = Math.floor(Math.random() * (i + 1)); var t = cells[i]; cells[i] = cells[j]; cells[j] = t; }
      var fog = "░▒░ ░";
      var grid = art.map(function (l) { var row = []; for (var x = 0; x < width; x++) row.push(fog[(x * 7 + l.length) % fog.length]); return row; });
      var shown = 0, total = cells.length, steps = 14, start = null;
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

    // Stat bars fill in.
    $$(".bar").forEach(function (b) { b.classList.add("pending"); });
    requestAnimationFrame(function () { setTimeout(function () { $$(".bar").forEach(function (b) { b.classList.remove("pending"); }); }, 150); });

    // Myth vs Movie: each card is a <details>, so clicking works without
    // us; r opens the lot, like the explorer.
    var myths = $$(".myth");
    var revealAll = function () { myths.forEach(function (m) { m.open = true; }); };
    var revealBtn = $("#reveal-all");
    if (revealBtn) revealBtn.addEventListener("click", revealAll);

    // Light the section tab for where you are.
    var tabs = $$(".section-tabs a");
    var sections = tabs.map(function (a) { return $(a.getAttribute("href")); });
    if ("IntersectionObserver" in window) {
      var io = new IntersectionObserver(function (entries) {
        entries.forEach(function (e) {
          if (e.isIntersecting) {
            tabs.forEach(function (a) { a.classList.toggle("active", a.getAttribute("href") === "#" + e.target.id); });
          }
        });
      }, { rootMargin: "-30% 0px -60% 0px" });
      sections.forEach(function (s) { if (s) io.observe(s); });
    }

    // The explorer's keys, for people who never left the terminal.
    var scrollMode = reduceMotion ? "auto" : "smooth";
    document.addEventListener("keydown", function (e) {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      var tag = (e.target.tagName || "").toLowerCase();
      if (tag === "input" || tag === "textarea") return;
      var prev = monster.getAttribute("data-prev"), next = monster.getAttribute("data-next");
      switch (e.key) {
        case "h": case "ArrowLeft": case "p": if (prev) location.href = prev; break;
        case "l": case "ArrowRight": case "n": if (next) location.href = next; break;
        case "g": case "Escape": location.href = "index.html"; break;
        case "r": revealAll(); break;
        case "j": window.scrollBy({ top: 80, behavior: scrollMode }); break;
        case "k": window.scrollBy({ top: -80, behavior: scrollMode }); break;
        default:
          if (e.key >= "1" && e.key <= "7") {
            // Tabs keep the explorer's numbers even when a monster has no
            // film or quotes, so find the one labelled with this key.
            var a = tabs.filter(function (t) { var k = $("kbd", t); return k && k.textContent === e.key; })[0];
            if (a) $(a.getAttribute("href")).scrollIntoView({ behavior: scrollMode });
          }
      }
    });
  }
})();
