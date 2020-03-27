CREATE TABLE test_users (
	user_id UUID NOT NULL UNIQUE DEFAULT uuid_generate_v4() PRIMARY KEY,
	username VARCHAR(32) CHECK (char_length(username) > 4) NOT NULL UNIQUE,
	password VARCHAR(64) NOT NULL,
	email VARCHAR(64) CHECK (char_length(email) > 5) NOT NULL UNIQUE,
	type VARCHAR(16) NOT NULL DEFAULT 'standart',
	first_name VARCHAR(32),
	last_name VARCHAR(32),
	gender VARCHAR(6),
	country VARCHAR(64),
	city VARCHAR(64),
	birth_date timestamp without time zone CHECK (birth_date > '1900-01-01')
);
