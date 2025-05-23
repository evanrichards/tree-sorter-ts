const config1 = {
  /** tree-sorter-ts: keep-sorted
      deprecated-at-end with-new-line **/
  alpha: "first",

  beta: "second",

  gamma: true,

  /** @deprecated */
  oldApi: "old",
},
}const config2 = {
  /** 
   * tree-sorter-ts: keep-sorted
   * deprecated-at-end 
   * with-new-line 
   **/
  alpha: "first",

  beta: "second",

  zebra: "last",

  /** @deprecated */
  oldFeature: false,
},
}const config3 = {
  /**
   * tree-sorter-ts: keep-sorted
   * deprecated-at-end with-new-line
   */
  alpha: "first",

  charlie: "third",

  delta: true,

  /** @deprecated Use newValue */
  oldValue: "old",
},
}const config4 = {
  /**
   * tree-sorter-ts: keep-sorted
   *   with-new-line
   *   deprecated-at-end
   */
  alpha: "first",

  beta: "second",

  zebra: "last",

  /** @deprecated */
  oldItem: "old",
};