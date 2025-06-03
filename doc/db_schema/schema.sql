DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    vip_level     INTEGER,
    maker_fee      REAL,
    taker_fee      REAL
);

DROP TABLE IF EXISTS balances;
CREATE TABLE balances
(
    user_id   TEXT,
    asset     TEXT,
    available REAL NOT NULL,
    locked    REAL NOT NULL,
    PRIMARY KEY (user_id, asset)
);

DROP TABLE IF EXISTS orders;
CREATE TABLE orders
(
    id             TEXT PRIMARY KEY,
    user_id        TEXT,
    market         TEXT,    -- ex: BTC/USDT ETH/USDT
    side           INTEGER, -- 0=Bid,1=Ask
    price          REAL,
    original_size  REAL,
    remaining_size REAL,
    quote_amount   REAL, -- only for market order
    avg_dealt_price REAL,
    type           INTEGER, -- 0=LIMIT,1=MARKET
    mode           INTEGER, -- 0=MAKER,1=TAKER
    status         TEXT,    -- NEW, FILLED, CANCELED
    created_at     DATETIME,
    updated_at     DATETIME
);


DROP TABLE IF EXISTS trades;
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