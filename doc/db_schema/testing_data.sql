delete from orders where TRUE;
delete from trades where TRUE;
update balances set available = 0, locked = 0 where TRUE;

-- Create Testing Maker Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee) values ('1', 'market_maker', '0x001', 7, 0.0, 0.02);
-- Create Balances for Maker Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('1', 'USDT', 10000000, 0),
       ('1', 'BTC', 100, 0),
       ('1', 'ETH', 10000, 0),
       ('1', 'DOT', 10000, 0);

