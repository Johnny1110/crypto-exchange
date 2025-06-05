delete from orders where TRUE;
delete from trades where TRUE;
update balances set available = 0, locked = 0 where TRUE;

-- Create Testing Maker Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('1', 'market_maker', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 7, 0.0001, 0.002);

-- Create Testing User Account
INSERT INTO users(id,username,password_hash,vip_level,maker_fee, taker_fee)
values ('U01_the_GOD', 'johnny', '$2a$10$z.kl4/Zazgme18gFCqwozOk5WoqMbhqAeZk5.zk55gwVgurQCwqpq', 1, 0.001, 0.002);

-- Create Balances for Maker Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('1', 'USDT', 10000000, 0),
       ('1', 'BTC', 100, 0),
       ('1', 'ETH', 10000, 0),
       ('1', 'DOT', 10000, 0);

-- Create Balances for Maker Account
INSERT INTO balances(user_id,asset,available,locked)
VALUES ('U01_the_GOD', 'USDT', 100000, 0),
       ('U01_the_GOD', 'BTC', 0, 0),
       ('U01_the_GOD', 'ETH', 10, 0),
       ('U01_the_GOD', 'DOT', 10, 0);

