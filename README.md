## Project Structure Expected
+ cmd/api-server/main.go - HTTP API

+ cmd/tcp-server/main.go - TCP sync

+ cmd/udp-server/main.go - UDP notify

+ cmd/grpc-server/main.go - gRPC

+ internal/auth/ - Auth logic

+ internal/manga/ - Manga management

+ internal/user/ - User management

+ internal/tcp/udp/websocket/grpc/ - Network protocols

+ pkg/models/ - Data structures

+ database/ - Utilities

+ proto/ - Protocol buffers

+ data/ - Manga/user JSON files

+ docs/ - Documentation

+ docker-compose.yml - DevOps setup​


## Tech Stack Overview
Programming Language: Go (Golang) for all services and tooling

Database: SQLite for persistent storage of users, manga, and progress

Cache (Bonus): Redis for caching frequently accessed manga data

Communication Protocols: TCP, UDP, HTTP (REST), gRPC, WebSocket—all implemented natively in Go

API Library: Gin (HTTP), Gorilla WebSocket, gRPC/Protobuf

Authentication: JWT for user sessions

Testing: Testify for unit/integration tests

Containerization/Deployment (Bonus): Docker, docker-compose for dev/test

Monitoring/Logging (Bonus): Basic logging; optional Prometheus/Grafana for metrics

External APIs: MangaDx for manga metadata and expansion