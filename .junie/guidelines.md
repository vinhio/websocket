# WebSocket Chat Application - Project Guidelines

## Project Overview
This project is a real-time chat application built with WebSockets. It consists of a Go backend server and an HTML/JavaScript frontend. The application allows multiple clients to connect to a central hub and exchange messages in real-time.

## Project Structure
- **Backend (Go)**:
  - `main.go`: Entry point that sets up the WebSocket endpoint using the gflydev/core framework
  - `websocket.go`: Implements the WebSocket handler and hub management
  - `hub.go`: Manages client connections and message broadcasting
  - `client.go`: Handles individual WebSocket client connections
  - `websocket/`: A package containing the WebSocket implementation

- **Frontend (HTML/JavaScript)**:
  - `public/index.html`: Simple chat interface
  - `public/websocket.js`: ReconnectingWebSocket implementation for reliable connections

## Development Guidelines
1. **Running the Application**:
   - Use `make run` to start the server locally
   - Access the chat interface at http://localhost:8080 (or the configured port)

2. **Building the Application**:
   - Use `make build` to build the application
   - Use `make release` to build for deployment

3. **SSL Certificates**:
   - Use `make certs` to generate SSL certificates for secure connections

4. **Code Style**:
   - Follow Go standard formatting (use `go fmt`)
   - Include comprehensive comments for public functions and methods
   - Maintain error handling throughout the codebase

5. **Testing**:
   - When making changes, test the WebSocket functionality by:
     - Opening multiple browser windows to ensure message broadcasting works
     - Testing reconnection by temporarily stopping and restarting the server
     - Verifying proper cleanup of disconnected clients

## Architecture
The application follows a hub-and-spoke architecture:
- A central Hub manages all client connections
- Each Client has read and write pumps running in separate goroutines
- Messages sent by one client are broadcast to all connected clients
- The frontend uses a ReconnectingWebSocket to maintain persistent connections

## Security Considerations
- Origin checking is implemented to prevent unauthorized connections
- Message size limits are enforced to prevent DoS attacks
- Proper connection timeouts and cleanup are implemented
