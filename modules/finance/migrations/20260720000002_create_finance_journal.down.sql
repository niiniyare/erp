-- Reverse: drop journal tables in reverse dependency order.

DROP TABLE IF EXISTS finance_ledger_entry;
DROP TABLE IF EXISTS finance_journal_entry_line;
DROP TABLE IF EXISTS finance_journal_entry;
DROP TABLE IF EXISTS finance_journal;
