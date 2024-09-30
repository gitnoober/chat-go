# Chat Application Overview

## Introduction
This chat application server is built using Go and provides WebSocket connectivity for real-time communication between users. The application includes features such as user signup, JWT authentication, and a message routing system. It was created for learning purposes. :alien:

## TODO :art:
- Add UTs.

## Features

### WebSocket Connection
- Users can connect to the server using the WebSocket protocol.
- Each client connection is managed through a `Pool` that tracks active users.

### User Authentication
- **JWT (JSON Web Tokens)** are used for secure user authentication.
- Tokens are validated on connection requests and provide a mechanism for refreshing session tokens.

### Message Handling
- Messages are routed from one user to another through the server.
- The message format follows `receiverID:message` for proper routing.

### Online User Management
- An endpoint is available to fetch all currently connected users.
- The application keeps track of users in the connection pool.

### Refresh Token Flow
- The application supports a refresh token mechanism to allow users to obtain new access tokens without re-authenticating.
- Refresh tokens are stored in Redis for efficient retrieval and management.

### Profile Picture Generation
- The application can generate random profile picture URLs using Gravatar and integrates with Unsplash for fetching random avatars.

## Concurrency and Rate Limiting
- The server handles concurrent requests using goroutines, ensuring that multiple clients can connect and communicate simultaneously.
- Rate limiting is implemented to control the number of requests a user can make, preventing abuse.

## Technology Stack
- **Programming Language**: Go
- **Web Framework**: `net/http` for HTTP server and `github.com/coder/websocket` for WebSocket handling
- **Database**: MySQL for user data
- **Token Management**: JWT for authentication and Redis for refresh token storage
- **Profile Picture Services**: Gravatar and Unsplash APIs

## Conclusion
This chat application provides a robust foundation for real-time communication, focusing on security and scalability. Future improvements could include more advanced message handling, user presence indicators, and enhanced security features.