// Test file for with-new-line configuration

// Default behavior - no extra newlines
const obj1 = {
  /** tree-sorter-ts: keep-sorted */
  apple: 2,
  banana: 3,
  cherry: 4,
  zebra: 1,
};

// With new line configuration - adds extra newlines between entries
const obj2 = {
  /** tree-sorter-ts: keep-sorted with-new-line */
  apple: 2,

  banana: 3,

  cherry: 4,

  zebra: 1,
};

// Test with comments
const obj3 = {
  /** tree-sorter-ts: keep-sorted with-new-line */
  // Comment for apple
  apple: 2,

  banana: 3, // inline comment

  cherry: 4,

  // Comment for zebra
  zebra: 1,
};
