/* eslint-disable no-restricted-globals */
type RudderstackWindowSDK = {
  snippetExecuted?: boolean;

  setDefaultInstanceKey: Function;
  load: Function;
  ready: Function;
  page: Function;
  track: Function;
  identify: Function;
  alias: Function;
  group: Function;
  reset: Function;
  setAnonymousId: Function;
  startSession: Function;
  endSession: Function;
  consent: Function;
};

declare global {
  interface Window {
    RudderSnippetVersion: string;
    rudderanalytics: RudderstackWindowSDK;
    rudderAnalyticsBuildType: string;
    rudderAnalyticsAddScript: (url: string, extraAttributeKey?: string, extraAttributeVal?: string) => void;
    rudderAnalyticsMount: Function;
  }
}

export function rudderstackInit() {
  window.RudderSnippetVersion = '3.0.60';
  let identifier = 'rudderanalytics' as const;
  if (!window[identifier]) {
    // @ts-expect-error
    window[identifier] = [];
  }
  let rudderanalytics = window[identifier];

  if (Array.isArray(rudderanalytics)) {
    if (rudderanalytics.snippetExecuted === true && window.console && console.error) {
      console.error('RudderStack JavaScript SDK snippet included more than once.');
    } else {
      rudderanalytics.snippetExecuted = true;
      window.rudderAnalyticsBuildType = 'legacy';
      let sdkBaseUrl = 'https://cdn.rudderlabs.com';
      let sdkVersion = 'v3';
      let sdkFileName = 'rsa.min.js';
      let scriptLoadingMode = 'async';
      let methods = [
        'setDefaultInstanceKey',
        'load',
        'ready',
        'page',
        'track',
        'identify',
        'alias',
        'group',
        'reset',
        'setAnonymousId',
        'startSession',
        'endSession',
        'consent',
      ] as const;
      for (let i = 0; i < methods.length; i++) {
        let method = methods[i];
        rudderanalytics[method] = (function (methodName) {
          return function () {
            if (Array.isArray(window[identifier])) {
              rudderanalytics.push([methodName].concat(Array.prototype.slice.call(arguments)));
            } else {
              let _methodName;
              (_methodName = window[identifier][methodName]) === null ||
                _methodName === undefined ||
                _methodName.apply(window[identifier], arguments);
            }
          };
        })(method);
      }
      try {
        new Function(
          'class Test{field=()=>{};test({prop=[]}={}){return prop?(prop?.property??[...prop]):import("");}}'
        );
        window.rudderAnalyticsBuildType = 'modern';
      } catch (e) {}
      let head = document.head || document.getElementsByTagName('head')[0];
      let body = document.body || document.getElementsByTagName('body')[0];
      window.rudderAnalyticsAddScript = function (url, extraAttributeKey, extraAttributeVal) {
        let scriptTag = document.createElement('script');
        scriptTag.src = url;
        scriptTag.setAttribute('data-loader', 'RS_JS_SDK');
        if (extraAttributeKey && extraAttributeVal) {
          scriptTag.setAttribute(extraAttributeKey, extraAttributeVal);
        }
        if (scriptLoadingMode === 'async') {
          scriptTag.async = true;
        } else if (scriptLoadingMode === 'defer') {
          scriptTag.defer = true;
        }
        if (head) {
          head.insertBefore(scriptTag, head.firstChild);
        } else {
          body.insertBefore(scriptTag, body.firstChild);
        }
      };
      window.rudderAnalyticsMount = function () {
        (function () {
          if (typeof globalThis === 'undefined') {
            let getGlobal = function getGlobal() {
              if (typeof self !== 'undefined') {
                return self;
              }
              if (typeof window !== 'undefined') {
                return window;
              }
              return null;
            };
            let global = getGlobal();
            if (global) {
              Object.defineProperty(global, 'globalThis', {
                value: global,
                configurable: true,
              });
            }
          }
        })();
        window.rudderAnalyticsAddScript(
          ''
            .concat(sdkBaseUrl, '/')
            .concat(sdkVersion, '/')
            .concat(window.rudderAnalyticsBuildType, '/')
            .concat(sdkFileName),
          'data-rsa-write-key',
          '1vjCCxXFaLSCZL0JiIkR313ixXW'
        );
      };
      if (typeof Promise === 'undefined' || typeof globalThis === 'undefined') {
        window.rudderAnalyticsAddScript(
          'https://polyfill-fastly.io/v3/polyfill.min.js?version=3.111.0&features=Symbol%2CPromise&callback=rudderAnalyticsMount'
        );
      } else {
        window.rudderAnalyticsMount();
      }
      let loadOptions = {};
      rudderanalytics.load('1vjCCxXFaLSCZL0JiIkR313ixXW', 'https://grafana.dataplane.rudderstack.com', loadOptions);
    }
  }
}
