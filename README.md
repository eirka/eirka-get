# eirka-get

[![GoDoc](http://img.shields.io/badge/go-documentation-brightgreen.svg?style=flat-square)](https://godoc.org/github.com/eirka/eirka-get)
[![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg?style=flat-square)](https://golang.org/doc/devel/release.html)
[![Build Status](https://img.shields.io/badge/build-passing-green.svg?style=flat-square)](https://github.com/eirka/eirka-get)

## Overview

eirka-get is the read-only API service for the Eirka imageboard system. It serves JSON responses for threads, posts, images, tags, and other imageboard features.

## Features

- **Caching System**: Efficient Redis caching with circuit breaker pattern for resilience
- **Singleflight Pattern**: Prevents thundering herd problems by deduplicating concurrent requests
- **JWT Authentication**: Secure authentication for user-specific features
- **Analytics**: Request tracking and performance monitoring
- **Efficient Pagination**: Paginated results for large data sets

## Architecture

The project follows a clean MVC architecture:

- **Controllers**: HTTP handlers that process requests and return JSON responses
- **Models**: Database interaction layer that fetches data from MySQL
- **Middleware**: Request processing components (caching, authentication, analytics)
- **Config**: Application configuration management

## Cache System

The caching middleware implements Redis caching with several advanced features:

1. **Circuit Breaker Pattern**: Automatically detects Redis failures and bypasses cache when Redis is experiencing issues
2. **Cache Key Management**: Organizes cache keys by resource type
3. **Singleflight Pattern**: Prevents duplicate database queries for concurrent requests to the same resource
4. **Intelligent Caching**: Caches only appropriate endpoints and skips dynamic queries

## Endpoints

The API provides the following main endpoints:

- **Threads**: `/thread/:imageboard/:thread/:page`
- **Posts**: `/post/:imageboard/:post`
- **Images**: `/image/:imageboard/:image`
- **Tags**: `/tag/:imageboard/:tag`, `/tags/:imageboard`
- **Search**: `/tagsearch/:imageboard/:search`, `/threadsearch/:imageboard/:search`
- **Favorites**: `/favorites/:imageboard`, `/favorite/:imageboard/:thread`, `/favorited/:imageboard/:thread`
- **Directory**: `/directory/:imageboard`
- **New Content**: `/new/:imageboard`
- **Popular Content**: `/popular/:imageboard`

## Installation

```bash
# Clone the repository
git clone https://github.com/eirka/eirka-get.git

# Install dependencies
go mod tidy

# Build the application
go build
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific tests
go test -v ./middleware -run TestCache

# Run tests with coverage
go test ./... -cover
```

## Contributing

1. Follow Go standard formatting (gofmt)
2. Group imports in 3 blocks: standard lib, third-party, project-local
3. Ensure tests pass before submitting changes
4. Use error wrapping with `SetMeta()` to provide context

## License

See [LICENSE](LICENSE) file for details.
