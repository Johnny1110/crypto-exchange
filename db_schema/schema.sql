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
    id             TEXT PRIMARY KEY,
    user_id        TEXT,
    market         TEXT,    -- ex: BTC/USDT ETH/USDT
    side           INTEGER, -- 0=Bid,1=Ask
    price          REAL,
    original_size  REAL,
    remaining_size REAL,
    quote_amount   REAL,
    type           INTEGER, -- 0=LIMIT,1=MARKET
    mode           INTEGER, -- 0=MAKER,1=TAKER
    status         TEXT,    -- NEW, FILLED, CANCELED
    created_at     DATETIME,
    updated_at     DATETIME
);

create table trades
(
    id           INTEGER
        primary key autoincrement,
    ask_order_id TEXT     not null,
    bid_order_id TEXT     not null,
    price        REAL     not null,
    size         REAL     not null,
    timestamp    DATETIME not null
);