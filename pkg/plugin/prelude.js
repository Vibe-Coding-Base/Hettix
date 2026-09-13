// Prelude injected into every plugin runtime before the plugin's own code. It
// builds the `hettix` API on top of the minimal primitives the host exposes as
// `__host`, and provides `__dispatch`, which the host calls per proxied response.
(function (global) {
  var host = global.__host;
  var responseHandlers = [];

  function str(v) {
    return v === undefined || v === null ? "" : String(v);
  }

  function isJavaScript(resp) {
    var ct = str(resp.contentType).toLowerCase();
    if (ct.indexOf("javascript") >= 0 || ct.indexOf("ecmascript") >= 0) {
      return true;
    }
    var path = str(resp.url).toLowerCase().split("?")[0];
    return /\.(m?js)$/.test(path);
  }

  global.hettix = {
    // plugin declares metadata: { id, name, description, version, capabilities,
    // defaultEnabled }.
    plugin: function (meta) {
      host.register(meta || {});
    },

    // onResponse registers a callback invoked for each proxied response.
    onResponse: function (fn) {
      if (typeof fn !== "function") {
        throw new Error("hettix.onResponse requires a function");
      }
      responseHandlers.push(fn);
    },

    // addEndpoint records a discovered endpoint. Accepts either a host and path,
    // or a single absolute URL.
    addEndpoint: function (hostOrURL, path) {
      if (path === undefined) {
        var m = /^https?:\/\/([^/]+)(\/[^\s]*)?/i.exec(str(hostOrURL));
        if (m) {
          host.addEndpoint(m[1], m[2] || "/");
        }
        return;
      }
      host.addEndpoint(str(hostOrURL), str(path));
    },

    // addFinding records a finding: { title, description, severity, requestLogId }.
    addFinding: function (finding) {
      finding = finding || {};
      host.addFinding(
        str(finding.title),
        str(finding.description),
        str(finding.severity || "info"),
        finding.requestLogId || null
      );
    },

    log: function (message) {
      host.log(str(message));
    },
  };

  // Called by the host for each proxied response.
  global.__dispatch = function (resp) {
    resp.isJavaScript = function () {
      return isJavaScript(resp);
    };
    for (var i = 0; i < responseHandlers.length; i++) {
      responseHandlers[i](resp);
    }
  };

  global.__hasResponseHandlers = function () {
    return responseHandlers.length > 0;
  };
})(this);
