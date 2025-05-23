const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  activeFeature: true,
  beta: "test",
  /** @deprecated Use newApiUrl instead */
  oldApiUrl: "https://old.example.com",
  zebra: false,
  /** @deprecated Will be removed in v2.0 */  
  legacyMode: true,
  alpha: "first",
  newApiUrl: "https://api.example.com", // replaces oldApiUrl
};

const anotherConfig = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  gamma: 123,
  oldSetting: "old", // @deprecated use newSetting
  delta: true,
  epsilon: "test",
  newSetting: "new",
};