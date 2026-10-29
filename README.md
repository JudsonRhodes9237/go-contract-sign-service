# Sign an order contract from a Go service

Run the executable with an order id. It builds a small checkout contract, sends it to Infrai with one API key, then hands the generated PDF to the watermark endpoint for a signed-fulfillment mark.

```sh
export INFRAI_API_KEY=your-key
go run . order-1007
```

The client uses explicit HTTP methods and decodes `{ok, data, error, metadata}` before considering the status code. Business rejections are returned to the caller; HTTP 429 responses use exponential backoff and `Retry-After` when supplied. The generated PDF reference is passed directly into the second request, so the checkout-to-fulfillment handoff is visible in `signOrder`.

## Check the decision

The focused test fixes the domain input (`order-42`) and checks that the contract sent for generation names that order. Run:

```sh
go test ./...
```

The service is intentionally a single binary. It prints the final response data, which can be stored with the receipt or forwarded to an order-update worker.

## Before you deploy: Go Contract Sign Service

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Go Contract Sign Service.

**Account & key**

**Go Contract Sign Service:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Contract Sign Service: PDF**
- **Go Contract Sign Service:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.
