# encproc

encproc is a homomorphic encryption library wrapper built in Go (currently with the lattigo library https://github.com/tuneinsight/lattigo). It leverages homomorphic encryption (currently using Lattigo's BGV scheme) to securely process encrypted data streams and integrates with a MySQL database. The project is containerized using Docker. 

## Features

| Feature                  | Description                                                                 |
|--------------------------|-----------------------------------------------------------------------------|
| Secure Aggregation       | Utilizes homomorphic encryption (BGV scheme) for secure data aggregation.   |
| RESTful API              | Provides endpoints to create streams, contribute data, and retrieve aggregates.|
| JWT Authentication       | Protects sensitive endpoints using JWT tokens.                              |
| Database Integration     | Uses MySQL for storing aggregation parameters and results.                  |
| Containerized Deployment | Docker and Docker Compose configurations for easy deployment.               |

## Architecture

| Module                     | Description                                                                                                    |
|----------------------------|----------------------------------------------------------------------------------------------------------------|
| Aggregator (`aggregator.go`) | Handles initialization and secure aggregation of ciphertexts. Uses a mutex to address concurrency issues.   |
| Database (`database.go`)     | Manages CRUD operations and schema initialization for aggregation parameters and results.                      |
| HTTP Handlers (`handlers.go`) | Implements API endpoints for creating streams, contributing data, and retrieving aggregates.                  |
| Middleware (`middleware.go`)  | Implements JWT authentication and request logging middleware.                                              |
| Utilities (`helpers.go`)      | Contains helper functions for JSON responses, environment configuration, and error logging.                   |
| Routing (`routes.go`)         | Defines API routes. Consider upgrading to a more flexible router for dynamic routes in the future.             |
| Client-side WASM             | Provides WebAssembly (WASM) files to perform encryption, decryption, and encoding on the client side.                     |

## Getting Started

### Prerequisites

Ensure you have the following installed:
- **Go 1.18+**
- **MySQL 8.0+**
- **Docker & Docker Compose**

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/collapsinghierarchy/encproc.git
   cd encproc
