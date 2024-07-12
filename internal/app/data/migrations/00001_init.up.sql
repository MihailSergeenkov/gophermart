BEGIN TRANSACTION;

CREATE TABLE users(
	id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	login VARCHAR(200) NOT NULL,
	password VARCHAR(300) NOT NULL
);
CREATE UNIQUE INDEX login_index ON users(login);

CREATE TABLE balance(
	user_id INT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
  current INT DEFAULT 0 NOT NULL,
  withdrawn INT DEFAULT 0 NOT NULL
);
CREATE UNIQUE INDEX user_id_index ON balance(user_id);

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders(
	number VARCHAR(200) NOT NULL,
	user_id INT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
  status order_status DEFAULT 'NEW' NOT NULL,
  accrual INT DEFAULT 0 NOT NULL,
  uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);
CREATE UNIQUE INDEX number_index ON orders(number);


CREATE TABLE withdrawals(
	order_number VARCHAR(200) NOT NULL,
	user_id INT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
  sum INT NOT NULL,
  processed_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);
CREATE UNIQUE INDEX order_number_index ON withdrawals(order_number);

COMMIT;
