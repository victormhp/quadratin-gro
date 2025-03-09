PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXIST categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXIST news (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    date TEXT NOT NULL,
    url TEXT NOT NULL,
    image TEXT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
           

-- .import --csv --skip 1 data/categories.csv categories
-- .import --csv --skip 1 data/news.csv news
