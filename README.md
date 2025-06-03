# crypto-exchange

<br>

---

A crypto-exchange implement by Golang

<br>

---

<br>

## startup ganache (A quick local Ethereum blockchain node)

### node.js is required

<br>

```
$> node install
$> node node_modules/.bin/ganache
```

<br>
<br>

## Document

* How to startup?

    ```
    > go install
    > go run
    ```

<br>

* Database

    * [Schema](doc/db_schema/db_schema_sql)
    * [Testing Data](doc/db_schema/testing_data.sql)

<br>

* [API Doc](doc)

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
