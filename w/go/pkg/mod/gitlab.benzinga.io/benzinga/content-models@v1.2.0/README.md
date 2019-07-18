[![pipeline status](https://gitlab.benzinga.io/benzinga/content-models/badges/master/pipeline.svg)](https://gitlab.benzinga.io/benzinga/content-models/commits/master)


# Content Models
This repository contains models for Benzinga content.

## Getting Started
Content Models is built in Go. To get started with Go, check out our [Go Basics](https://benzinga.atlassian.net/wiki/spaces/DEV/pages/112497925/Go+Basics) article on Confluence.

Effort to move toward using version 2 (`models.v2`) is ongoing and as of this writing nothing is using `models.v2`. It is a revision over the current models in Content Engine.

Services that publish or consume Benzinga content use the `models` package. This includes `feed-engine`, `pro-backend`, `sec-engine`, `newsdesk` and other services.

## Testing
```bash
make get test
```
