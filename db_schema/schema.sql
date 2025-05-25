CREATE TABLE users
(
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    token         TEXT UNIQUE
);

CREATE TABLE balances
(
    user_id   TEXT,
    asset     TEXT,
    available REAL NOT NULL,
    locked    REAL NOT NULL,
    PRIMARY KEY (user_id, asset)
);

CREATE TABLE orders
(
    id            TEXT PRIMARY KEY,
    user_id       TEXT,
    side          INTEGER, -- 0=Bid,1=Ask
    price         REAL,
    original_size  REAL,
    remaining_size REAL,
    type          INTEGER, -- 0=Maker,1=Taker,2=Market
    status        TEXT,    -- NEW, FILLED, CANCELED
    created_at    DATETIME,
    updated_at    DATETIME
);

CREATE TABLE trades
(
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id         TEXT,
    counter_order_id TEXT,
    price            REAL,
    size              REAL,
    timestamp        DATETIME
);