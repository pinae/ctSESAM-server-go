CREATE TABLE IF NOT EXISTS `domains` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `userid` TEXT NOT NULL,
  `created` INTEGER(4) DEFAULT (DATETIME('now', 'localtime')),
  `data` BLOB
);
CREATE INDEX  IF NOT EXISTS `userid_idx` ON `domains` (`userid`);