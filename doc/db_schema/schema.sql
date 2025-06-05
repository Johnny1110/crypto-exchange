DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    vip_level     INTEGER DEFAULT 1,
    maker_fee      REAL NOT NULL,
    taker_fee      REAL NOT NULL
);

DROP TABLE IF EXISTS balances;
CREATE TABLE balances
(
    user_id   TEXT NOT NULL,
    asset     TEXT NOT NULL,
    available REAL DEFAULT 0,
    locked    REAL DEFAULT 0,
    PRIMARY KEY (user_id, asset)
);

CREATE INDEX idx_balances_asset ON balances(asset);

DROP TABLE IF EXISTS orders;
CREATE TABLE orders
(
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL ,
    market         TEXT NOT NULL ,    -- ex: BTC/USDT ETH/USDT
    side           INTEGER NOT NULL, -- 0=Bid,1=Ask
    price          REAL DEFAULT 0,
    original_size  REAL DEFAULT 0,
    remaining_size REAL DEFAULT 0,
    quote_amount   REAL DEFAULT 0, -- only for market order
    avg_dealt_price REAL DEFAULT 0,
    type           INTEGER NOT NULL, -- 0=LIMIT,1=MARKET
    mode           INTEGER NOT NULL, -- 0=MAKER,1=TAKER
    status         TEXT NOT NULL,    -- NEW, FILLED, CANCELED
    fee_rate       REAL DEFAULT 0,
    fees           REAL DEFAULT 0,
    fee_asset      TEXT,
    created_at     DATETIME NOT NULL,
    updated_at     DATETIME NOT NULL
);

CREATE INDEX idx_orders_user_id ON orders(user_id, status, created_at);
-- Active orders by market, status (critical for order book)
CREATE INDEX idx_orders_market_side_status ON orders(market, status, created_at);
-- Time-based queries (recent orders, order history)
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_updated_at ON orders(updated_at);


DROP TABLE IF EXISTS trades;
create table trades
(
    id           INTEGER
        PRIMARY KEY AUTOINCREMENT,
    market       TEXT     not null,
    ask_order_id TEXT     NOT NULL,
    bid_order_id TEXT     NOT NULL,
    price        REAL     NOT NULL,
    size         REAL     NOT NULL,
    bid_fee_rate REAL,
    ask_fee_rate REAL,
    timestamp    DATETIME NOT NULL
);

-- Individual order trade lookup
CREATE INDEX idx_trades_ask_order_id ON trades(ask_order_id);
CREATE INDEX idx_trades_bid_order_id ON trades(bid_order_id);
-- Time-based trade queries
CREATE INDEX idx_trades_timestamp ON trades(market, timestamp);
-- Price-based queries (for analytics)
CREATE INDEX idx_trades_price ON trades(market, price);