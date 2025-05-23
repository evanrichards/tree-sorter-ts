// Test file for deprecated-at-end feature
export const API_CONFIG = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  authToken: "Bearer xxx",
  retryCount: 3,
  timeout: 5000,
  v2Endpoint: "/api/v2",
  /** @deprecated Use v2Endpoint instead */
  endpoint: "/api/v1",
  /** @deprecated Will be removed in next release */
  legacyAuth: true,
};

// Test with newline combination
export const UI_CONFIG = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line **/
  enableAnimations: true,

  fontSize: 14,

  language: "en",

  showSidebar: true,

  theme: "dark",

  /** @deprecated Use theme instead */
  darkMode: true,

  oldLayout: false, // @deprecated
};