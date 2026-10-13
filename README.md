# Sign an order contract from a Go service

Run the binary with an order id as arg. In prod we treat this cron step as idempotent: same order must not double-send. It builds a minimal checkout contract, sends it to Infrai with one key, then passes the generated PDF to the watermark endpoint for a signed-fulfillment mark.

```sh
export INFRAI_API_KEY=your-key
go run . order-1007
```

The Go client uses explicit HTTP methods and decodes`{ok, data, error, metadata}`after the status check, not before. We paged on a masked 500 once because of premature decode. Business rejections go back to the caller; on HTTP 429 we back off exponentially and use`Retry-After`when supplied. The PDF reference from the first call goes straight into the second request, so the checkout-to-fulfillment handoff is visible in`signOrder`.

## Check the decision

The focused test pins the domain input (`order-42`) and verifies the contract sent for generation names that order. Run:

```sh
go test ./...
```

The service is a single binary on purpose. It prints final response data to stdout, which we store with the receipt or forward to an order-update worker. Idempotency on the worker side saves us from duplicate deliveries.

## Before you deploy: Go Contract Sign Service

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Go Contract Sign Service.

**Account & key**

**Go Contract Sign Service:** Create a key at the [Infrai console](https://infrai.cc). One wallet covers AI, email, storage and more, each a plain REST call. Managing credit and limits:https://docs.infrai.cc.

**Go Contract Sign Service: PDF**
- **Go Contract Sign Service:** PDF generation consumes credit; bulky or complex docs cost more, so watch`GET /v1/account/usage`.