// JS Miner scans JavaScript responses for two things a pentester cares about:
// endpoints referenced by the script (added to the sitemap) and secrets that
// should never ship to the client (recorded as findings). Edit the rule tables
// below to tune coverage for your target.

hettix.plugin({
  id: "js-miner",
  name: "JS Miner",
  description:
    "Scans JavaScript responses for referenced endpoints (added to the sitemap) and leaked secrets (recorded as findings).",
  version: "2.0.0",
  capabilities: ["passive-scan"],
  defaultEnabled: true,
});

var MAX_ENDPOINTS = 500;
var MAX_SECRETS = 100;

// Secret matchers, ordered specific-first. Each: label, severity, and a regexp.
// Patterns follow the conventions used by gitleaks / trufflehog and Burp's own
// JS scanner so coverage matches what the industry expects.
var SECRET_RULES = [
  { label: "AWS access key ID", severity: "high", re: /\b(?:AKIA|ASIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|A3T[A-Z0-9])[A-Z0-9]{16}\b/g },
  { label: "AWS secret access key", severity: "high", re: /(?:aws.{0,20})?(?:secret|access).{0,20}["'][A-Za-z0-9/+]{40}["']/gi },
  { label: "Google API key", severity: "high", re: /\bAIza[0-9A-Za-z\-_]{35}\b/g },
  { label: "Google OAuth access token", severity: "high", re: /\bya29\.[0-9A-Za-z\-_]{20,}\b/g },
  { label: "Firebase Cloud Messaging key", severity: "high", re: /\bAAAA[A-Za-z0-9_-]{7}:[A-Za-z0-9_-]{140}\b/g },
  { label: "Slack token", severity: "high", re: /\bxox[baprs]-[0-9A-Za-z-]{10,72}\b/g },
  { label: "Slack webhook", severity: "high", re: /https:\/\/hooks\.slack\.com\/services\/T[A-Za-z0-9_]+\/B[A-Za-z0-9_]+\/[A-Za-z0-9_]+/g },
  { label: "GitHub token", severity: "high", re: /\b(?:ghp|gho|ghu|ghs|ghr)_[0-9A-Za-z]{36}\b/g },
  { label: "GitHub fine-grained token", severity: "high", re: /\bgithub_pat_[0-9A-Za-z_]{82}\b/g },
  { label: "GitLab personal access token", severity: "high", re: /\bglpat-[0-9A-Za-z\-_]{20}\b/g },
  { label: "Stripe secret key", severity: "high", re: /\b(?:sk|rk)_live_[0-9A-Za-z]{24,}\b/g },
  { label: "Square access token", severity: "high", re: /\b(?:sq0atp|sq0csp)-[0-9A-Za-z\-_]{22,43}\b/g },
  { label: "Twilio API key", severity: "high", re: /\bSK[0-9a-fA-F]{32}\b/g },
  { label: "SendGrid API key", severity: "high", re: /\bSG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}\b/g },
  { label: "Mailgun API key", severity: "high", re: /\bkey-[0-9a-f]{32}\b/g },
  { label: "Mailchimp API key", severity: "high", re: /\b[0-9a-f]{32}-us[0-9]{1,2}\b/g },
  { label: "npm access token", severity: "high", re: /\bnpm_[0-9A-Za-z]{36}\b/g },
  { label: "OpenAI API key", severity: "high", re: /\bsk-(?:proj-)?[0-9A-Za-z]{20,}\b/g },
  { label: "Cloudinary credentials", severity: "high", re: /cloudinary:\/\/[0-9]+:[0-9A-Za-z\-_]+@[0-9A-Za-z\-_]+/g },
  { label: "Private key", severity: "high", re: /-----BEGIN (?:RSA |EC |OPENSSH |DSA |PGP )?PRIVATE KEY(?: BLOCK)?-----/g },
  { label: "Basic auth credentials in URL", severity: "medium", re: /\bhttps?:\/\/[^\s:@/]+:[^\s:@/]+@[^\s/]+/g },
  { label: "Hard-coded bearer token", severity: "medium", re: /authorization["'\s]*[:=]["'\s]*bearer\s+[0-9A-Za-z._\-]{12,}/gi },
  { label: "JSON Web Token", severity: "medium", re: /\beyJ[0-9A-Za-z_\-]{8,}\.eyJ[0-9A-Za-z_\-]{8,}\.[0-9A-Za-z_\-]{8,}\b/g },
  {
    label: "Hard-coded credential",
    severity: "low",
    re: /(?:api[_-]?key|apikey|secret|client[_-]?secret|access[_-]?token|auth[_-]?token|passwd|password)["'\s]*[:=]["'\s]*["'][0-9A-Za-z._\-/+=]{12,}["']/gi,
  },
];

// Obvious documentation placeholders that would otherwise create noise. Kept
// deliberately narrow so real keys (which may contain short runs like "1234")
// are not discarded.
var PLACEHOLDER_RE = /(?:your[_-]?|example|placeholder|changeme|redacted|dummy|sample|foobar|xxxxxxxx|<[^>]+>)/i;

// Endpoints referenced from the script.
var QUOTED_PATH_RE = /["'`](\/[A-Za-z0-9_~\-./]+(?:\?[A-Za-z0-9_~\-.=&%/]*)?)["'`]/g;
var ABSOLUTE_URL_RE = /https?:\/\/[A-Za-z0-9.\-]+(?::\d+)?\/[A-Za-z0-9_~\-./?=&%#]*/g;

hettix.onResponse(function (resp) {
  if (!resp.isJavaScript()) {
    return;
  }

  var body = resp.body || "";
  mineEndpoints(resp, body);
  mineSecrets(resp, body);
});

function mineEndpoints(resp, body) {
  var seen = {};
  var count = 0;
  var m;

  QUOTED_PATH_RE.lastIndex = 0;
  while (count < MAX_ENDPOINTS && (m = QUOTED_PATH_RE.exec(body)) !== null) {
    if (resp.host && add(seen, resp.host, m[1])) {
      count++;
    }
  }

  ABSOLUTE_URL_RE.lastIndex = 0;
  while (count < MAX_ENDPOINTS && (m = ABSOLUTE_URL_RE.exec(body)) !== null) {
    var parsed = /^https?:\/\/([^/]+)(\/[^\s]*)?/i.exec(m[0]);
    if (parsed && add(seen, parsed[1], parsed[2] || "/")) {
      count++;
    }
  }
}

function add(seen, host, path) {
  var key = host + " " + path;
  if (seen[key]) {
    return false;
  }
  seen[key] = true;
  hettix.addEndpoint(host, path);
  return true;
}

function mineSecrets(resp, body) {
  var seen = {};
  var count = 0;

  for (var i = 0; i < SECRET_RULES.length && count < MAX_SECRETS; i++) {
    var rule = SECRET_RULES[i];
    var m;
    rule.re.lastIndex = 0;
    while ((m = rule.re.exec(body)) !== null) {
      var match = m[0];
      if (seen[match] || PLACEHOLDER_RE.test(match)) {
        continue;
      }
      seen[match] = true;
      count++;
      report(resp, rule, match);
      if (count >= MAX_SECRETS) {
        break;
      }
      // Zero-length or global-flag safety: advance past empty matches.
      if (m.index === rule.re.lastIndex) {
        rule.re.lastIndex++;
      }
    }
  }
}

function report(resp, rule, match) {
  var description =
    "Plugin JS Miner detected a " +
    rule.label +
    " in a JavaScript response.\n\n" +
    "Type:     " +
    rule.label +
    "\n" +
    "Location: " +
    resp.url +
    "\n\n" +
    "Match:\n" +
    match;

  hettix.addFinding({
    title: rule.label + " exposed in JavaScript",
    description: description,
    severity: rule.severity,
    requestLogId: resp.requestLogId,
  });
}
