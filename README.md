# stripe-proxy

A proxy for Stripe which allows the administrator to grant permission-restricting credentials. Actual Stripe credentials are only used for signing the generated credentials, and are never shared with the end consumer. In this way the proxy can never be accidentally skipped.

## Usage

There are two subcommands which can be used to interact with stripe-proxy. The `sign` command, which generates signed restricted credentials, and the `serve` command which runs the HTTP reverse proxy. All commands require the use of a stripe key with the `--key` parameter. Alternatively, the `generate` command will run web browser-accessible user interface for securely generating signed credentials.

### Generate

To start the client user interface webserver, use a command like the following:

```
stripe-proxy --port <custom_port> generate
```

The default port is `9090`. Point your web browser of choice at the root address (ex. `localhost:9090/`) and you will see the user interface.

### Serve

To start the reverse proxy, use a command like the following:

```
stripe-proxy --stripekey <your_stripe_private_key> serve
```

### Sign

To generate a set of signed credentials, you must first calculate the permissions vector as a uint32, and then pass that to the sign command. The vector is comprised of individual permissions flags corresponding to Stripe top level resources and whether you want to grant read, write, both, or none. You can run the sign command as follows:

```
# Grants read only access to /customer/ paths
stripe-proxy --stripekey <your_stripe_private_key> sign --input 64
```

Output:

```
sign called with stripeKey: sk_test_redacted and input 1000000
credentials: AAAAQA_<signature>
```

#### Calculation of bit offsets

The calculation for which bit corresponds to what is as follows:

Access:

```go
type Access int
const (
	None      = 0
	Read      = 1
	Write     = 2
	ReadWrite = 3
)
```

Resources:

```go
type StripeResource int
const (
	ResourceAll               StripeResource = 0
	ResourceBalance                          = 1
	ResourceCharges                          = 2
	ResourceCustomers                        = 3
	ResourceDisputes                         = 4
	ResourceEvents                           = 5
	ResourceFileUploads                      = 6
	ResourceRefunds                          = 7
	ResourceTokens                           = 8
	ResourceTransfers                        = 9
	ResourceTransferReversals                = 10
)
```

Individual bit mask:

```go
func resourceMask(resource StripeResource, access Access) uint32 {
	return uint64(access << (uint64(resource) * 2))
}
```

The simplest and **default** case of `1` corresponds to granting read only access to everything.
