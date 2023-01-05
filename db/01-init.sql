CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
);

INSERT INTO expenses ("title", "amount","note", "tags") VALUES ('coke', 10.5, 'test note', '{test-tags}')