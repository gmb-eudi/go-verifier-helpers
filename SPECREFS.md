# Specification references

These packages define product-internal wire/cache contracts; they do not
implement an external normative specification of their own. Where a contract
aligns with a public standard, the relevant one is:

| Package | Aligned standard |
|---|---|
| `handoffwire` | OpenID4VP (the `response_code` one-time redemption model) |
| `trustcache` | EUDI Architecture Reference Framework — trust-anchor usage (per-type, per-territory trusted lists) |

The exact wire shape and key layout ARE the contract; treat them as stable and
version any change.
