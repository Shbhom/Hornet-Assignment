CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price REAL NOT NULL
);

INSERT INTO products (name, price) VALUES
('Laptop', 999.99),
('Smartphone', 699.99),
('Headphones', 149.99),
('Keyboard', 79.99),
('Monitor', 199.99),
('Mouse', 29.99),
('Tablet', 299.99),
('Camera', 499.99),
('Speaker', 89.99),
('Smartwatch', 199.99);
