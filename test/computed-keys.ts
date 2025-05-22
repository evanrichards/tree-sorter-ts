enum StatusEnum {
  PENDING = "PENDING",
  ACTIVE = "ACTIVE",
  COMPLETED = "COMPLETED",
}

const handlers = {
  /** tree-sorter-ts: keep-sorted **/
  [StatusEnum.ACTIVE]: handleActive, // primary handler,
  [StatusEnum.COMPLETED]: handleCompleted,
  [StatusEnum.PENDING]: handlePending,
};

function handlePending() {}
function handleActive() {}
function handleCompleted() {}