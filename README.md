# crypto-exchange

<br>

---

A crypto-exchange implement by Golang

<br>

---

## Tutorial Resource:

How to create a cryptocurrency exchange from scratch -- Anthony GG

<br>

## startup ganache (A quick local Ethereum blockchain node)

### node.js is required

<br>

```
$> node install
$> node node_modules/.bin/ganache
```

<br>

## Log:

2025/04/05 [EP1](https://youtu.be/5r1wHkmb3HM?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=1351): 22:30

2025/04/12 [EP2](https://youtu.be/SvitKOkJmm8?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=1199): 20:00

2025/04/13 [EP2](https://youtu.be/SvitKOkJmm8?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=5155s): 1:26:00

2025/04/19 [EP3](https://youtu.be/oE8TPzDIzLY?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=1851) 30:50

2025/04/20 [EP4](https://youtu.be/xpXT127JXEU?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=32) 00:30

2025/04/26 [EP4](https://youtu.be/xpXT127JXEU?list=PL0xRBLFXXsP5Q_a9FjmDfgtWatLHJVxGn&t=2893) 48:13 

2025/05/11 [EP5](https://youtu.be/bkzEohennvs?t=2018) 33:38 (Stuck into integrate ganache, can not do transfer token by go-ethereum wheels)

2025/05/17 [EP5](https://youtu.be/bkzEohennvs?t=3980) 1:06:20 (ETH transfer problem solved, and prepare to create user module, and integrate with Matches)

2025/05/18 [EP5](https://youtu.be/bkzEohennvs?t=8109) 2:15:09 (Token Transfer done, but don't know why to wallet address didn't received the ETH token)

2025/05/20 revamp engine-v2 by Chat-GPT [link](https://chatgpt.com/share/682c6180-9bb4-8003-bbd5-22332df49a82)

<br>
<br>

## Order Book Benchmark Test:

<br>

1. Limit Order (all maker)

```
goos: darwin
goarch: amd64
pkg: github.com/johnny1110/crypto-exchange/engine-v2/book
cpu: VirtualApple @ 2.50GHz
BenchmarkMakeLimitOrder
BenchmarkMakeLimitOrder-8   	 1978701/s	       694.7 ns/op
```

<br>

2. Limit Order full match (Taker)

```
goos: darwin
goarch: amd64
pkg: github.com/johnny1110/crypto-exchange/engine-v2/book
cpu: VirtualApple @ 2.50GHz
BenchmarkTakeLimitOrder_FullMatch
BenchmarkTakeLimitOrder_FullMatch-8   	 1882754/s	       743.3 ns/op
```

<br>

3. Market Order

```
goos: darwin
goarch: amd64
pkg: github.com/johnny1110/crypto-exchange/engine-v2/book
cpu: VirtualApple @ 2.50GHz
BenchmarkTakeMarketOrder
BenchmarkTakeMarketOrder-8   	 1789549/s	       680.6 ns/op
```

<br>

4. Cancel Order

```
goos: darwin
goarch: amd64
pkg: github.com/johnny1110/crypto-exchange/engine-v2/book
cpu: VirtualApple @ 2.50GHz
BenchmarkCancelOrder
BenchmarkCancelOrder-8   	 5864726/s	       259.0 ns/op
```
