BEGIN TRANSACTION;

DROP INDEX order_number_index;
DROP TABLE withdrawals;
DROP INDEX number_index;
DROP TABLE orders;
DROP TYPE order_status;
DROP INDEX login_index;
DROP TABLE users;

COMMIT;
