# encproc

**encproc** is a homomorphic encryption library wrapper built in Go, currently utilizing the [Lattigo](https://github.com/tuneinsight/lattigo) library. It leverages homomorphic encryption (currently using Lattigo's BGV scheme) to securely process encrypted data streams and integrates with a MySQL database. The project is containerized using Docker.

This framework can be described as a variation of "Encrypted Processing as a Service" (EPaaS). A key aspect of this concept is the "separation of duties" assumption, which ensures that the server processing encrypted data does not have access to the secret key and ideally even doesn't learn the decryption results. To achieve this, client-side components (compiled in WebAssembly format) are provided. These can be used programmatically outside the server's reach to produce ciphertexts, which the server processes by invoking appropriate APIs.

Currently, this project is in its early alpha stage and includes only minimal functionality. Its sole purpose for now is to provide sufficient tools for people with no cryptographic expertise to experiment and prototype. Therefore, if you are such a person, you should primarily be interested in the other [encproc-decryptor](https://github.com/collapsinghierarchy/encproc-decryptor) repository. The JavaScript functionality exposed by our WebAssembly-compiled modules provides sufficient tools to experiment with the full power of the engine without needing to worry about cryptographic configurations.

You can find examples of the usage of this API here -- [encproc-decryptor examples](https://github.com/collapsinghierarchy/encproc-decryptor/tree/main/client-side). The API is documented with swagger, see [Swagger Documentation](https://pseudocrypt.site/docs/).

> **Note:** You must have access to a running encproc engine (see [encproc](https://github.com/collapsinghierarchy/encproc)) and a valid JWT token for authentication to begin experimenting. If you do not wish to set up your own engine, contact encproc@gmail.com for connection details. If you want to try the client-side examples, you can do so immediately, as they are preconfigured with a running instance. Currently, the project is in a highly experimental state and is not yet ready for production use. Visit the [Encproc Homepage](https://pseudocrypt.site/) to view the roadmap.

## Roadmap

Our [roadmap](https://pseudocrypt.site/) includes expanding the engine with additional functionalities—such as standard statistical computations (e.g., mean, standard deviation, variance, correlations, matchings, and more)—and integrating further HE schemes, such as CKKS. We deliberately omit the prefix "fully" from Homomorphic Encryption because we focus on use cases with guaranteed efficiency. Although bootstrapping mechanisms—an inherently inefficient component when the functionality requires an exceedingly high multiplication depth—may be incorporated in the future, our experience shows that many use cases can be efficiently realized using "simple" (leveled) HE.

Currently, parameterizations are limited to one static, pre-defined parameter set:

```go
var def_parameter = {
    "LogN": 12,
    "LogQ": [58],
    "PlaintextModulus": 65537
}
```
This parameter set is neither the most secure nor the most efficient or fitting for every use case. The limitation of this parameter set being non-configurable at the client side during key generation will be lifted in the future. 

## Alternative EPaaS Frameworks

It should be noted that alternative EPaaS frameworks exist. Multi-key or threshold cryptosystems remove the need for a strict “separation of duties” assumption, but they introduce an operational constraint: users must be online simultaneously (or at specific stages) to collectively generate or combine keys for decryption. This can be impractical for certain web services that need asynchronous or on-demand decryption without waiting for all participants. So, if real-time user participation is acceptable, Multi-key or Threshold Encryption may indeed be preferable. If not, which we assume is the standard scenario an a web service setting, the “separation of duties” approach can offer a more practical alternative.

## Features

- **Secure Aggregation**: Employs homomorphic encryption (BGV scheme) for secure data aggregation.
- **RESTful API**: Offers endpoints to create streams, contribute data, and retrieve aggregates.
- **Database Integration**: Utilizes MySQL for storing encrypted aggregation parameters and results.
- **Containerized Deployment**: Provides Docker configurations for easy deployment.
- **Client-side WASM Components**: Supplies WebAssembly modules for client-side encryption, decryption, and encoding from a JS environment. Examples are shown here [encproc-decryptor](https://github.com/collapsinghierarchy/encproc-decryptor)

## Architecture

The project is structured into several key modules:

- **Aggregator (`aggregator.go`)**: Manages initialization and secure aggregation of ciphertexts.
- **Database (`database.go`)**: Handles CRUD operations and schema initialization for aggregation parameters and results.
- **HTTP Handlers (`handlers.go`)**: Implements API endpoints for stream management and data contribution.
- **Middleware (`middleware.go`)**: Provides authorization mechanisms for secure API access.
- **Utilities (`helpers.go`)**: Contains helper functions for JSON responses, environment configuration, and error logging.
- **Routing (`routes.go`)**: Defines API routes and associates them with corresponding handlers.

## Getting Started

### Prerequisites

Ensure you have the following installed:

- **Go 1.18+**
- **MySQL 8.0+**

or

- **Docker & Docker Compose**

### Installation

**Build and Run with Docker**:
Setup the environment configuration paramaters within the `docker-compose.yml` file. 

#### Configuration
Configuration settings for both the API and the database are managed through environment variables. Below is an overview of the primary variables used and their effects:

### MySQL Database Configuration (db service)
- **`MYSQL_ROOT_PASSWORD`**: Sets the root password for the MySQL database.
- **`MYSQL_DATABASE`**: Specifies the name of the default database created at container initialization.
- **`MYSQL_USER`**: Defines the username for a non-root user with access to the MySQL database.
- **`MYSQL_PASSWORD`**: Sets the password for the non-root MySQL user.

These variables ensure that the MySQL container is set up with the correct credentials and database structure.

### API Server Configuration (api service)
- **`DB_HOST`**: Specifies the hostname where the MySQL database is running (set to `db` by default in Docker, referring to the database container).
- **`DB_PORT`**: The port on which the MySQL database is exposed (default is `3306`).
- **`DB_NAME`**: The name of the MySQL database to which the API connects (should match `MYSQL_DATABASE`).
- **`DB_USER`**: The username for accessing the MySQL database (should match `MYSQL_USER`).
- **`DB_PASSWORD`**: The password for the MySQL database user (should match `MYSQL_PASSWORD`).
- **`SECRET_KEY`**: The secret key used for JWT authentication. (Note: Some documentation might refer to this as `JWT_SECRET`.)

#### Docker
Ensure Docker is installed on your system and is up and running. Then, execute:
```bash
docker compose --build
docker compose up
```
These commands will build the Docker images and start the services as defined in the `docker-compose.yml` file.

## Usage

If you want to prototype and experiment with this engine, you can either run it yourself or use an already running instance provided by someone else. The latter is the primary intended use case for this engine.

- **Running Locally**: If you are running the engine locally, you may disable the authorization middleware in the source code for easier experimentation.
- **Using a Remote Instance**: If you are experimenting with a running instance hosted by someone else, you will need to register a data stream with the engine.

For detailed instructions on how to use this engine, please refer to the client-side repository [encproc-decryptor](https://github.com/collapsinghierarchy/encproc-decryptor). Below, we provide an overview of the API endpoints.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m 'Add your feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a Pull Request.

Please ensure your code adheres to the project's coding standards and includes appropriate tests.

## License

This project is licensed under the Apache-2.0 License. See the [LICENSE](LICENSE) file for details.

- [encproc decryptor Repository](https://github.com/collapsinghierarchy/encproc-decryptor) -- The client side part that harnesses the powers of the encrypted processing as a service engine.
- [encproc Homepage](https://collapsinghierarchy.github.io/encproc-page/) -- The hompage of this project with a roadmap and an introduction to the overall encproc project.
- Contact: encproc@gmail.com

Happy encrypted processing!
