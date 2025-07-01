# encproc

**encproc** is a homomorphic encryption library wrapper built in Go, currently utilizing the [Lattigo](https://github.com/tuneinsight/lattigo) and [openfhe](https://github.com/openfheorg/openfhe-development) libraries. It leverages homomorphic encryption (currently using BGV scheme) to securely process encrypted data streams and integrates with a MySQL database. The project is containerized using Docker.

This framework can be described as a variation of "Encrypted Processing as a Service" (EPaaS). A key aspect of this concept is the "separation of duties" assumption, which ensures that the server processing encrypted data does not have access to the secret key and ideally even doesn't learn the decryption results. To achieve this, client-side components (compiled in WebAssembly format) are provided. These can be used programmatically outside the server's reach to produce ciphertexts, which the server processes by invoking appropriate APIs.

Currently, this project is in its early alpha stage and includes only minimal functionality. Its sole purpose for now is to provide sufficient tools for people with no cryptographic expertise to experiment and prototype. The JavaScript functionality exposed by our WebAssembly-compiled modules provides sufficient tools to experiment with the full power of the engine without needing to worry about cryptographic configurations.

The API is also documented with swagger, see [Swagger Documentation](https://pseudocrypt.site/docs/). Currently it's out of date. See into the routes.go for being up-to-date.

> **Note:** You must have access to a running encproc engine (see [encproc](https://github.com/collapsinghierarchy/encproc)). If you do not wish to set up your own engine, contact encproc@gmail.com for connection details. If you want to try the client-side examples, you can do so immediately, as they are preconfigured with a running instance. Currently, the project is in a highly experimental state and is not yet ready for production use. Visit the [Encproc Homepage](https://pseudocrypt.site/) to view the roadmap.

## Roadmap

Our [roadmap](https://pseudocrypt.site/) includes expanding the engine with additional functionalities—such as standard statistical computations (e.g., mean, standard deviation, variance, correlations, matchings, and more)—and integrating further HE schemes, such as CKKS. We deliberately omit the prefix "fully" from Homomorphic Encryption because we focus on use cases with guaranteed efficiency. Although bootstrapping mechanisms—an inherently inefficient component when the functionality requires an exceedingly high multiplication depth—may be incorporated in the future, our experience shows that many use cases can be efficiently realized using "simple" (leveled) HE.

Currently, parameterizations are limited to one static, pre-defined parameter set. This parameter set is neither the most secure nor the most efficient or fitting for every use case. The limitation of this parameter set being non-configurable at the client side during key generation will be lifted in the future. 

## Alternative EPaaS Frameworks

It should be noted that alternative EPaaS frameworks exist. Multi-key or threshold cryptosystems remove the need for a strict “separation of duties” assumption, but they introduce an operational constraint: users must be online simultaneously (or at specific stages) to collectively generate or combine keys for decryption. This can be impractical for certain web services that need asynchronous or on-demand decryption without waiting for all participants. So, if real-time user participation is acceptable, Multi-key or Threshold Encryption may indeed be preferable. If not, which we assume is the standard scenario an a web service setting, the “separation of duties” approach can offer a more practical alternative.

### Monitoring with Prometheus
`encproc` exposes **Prometheus** metrics so you can watch throughput, latency, and resource use while your encrypted workloads run.

```text
Endpoint:  /metrics           (enabled by default)
Port:      9000               
```
Currently all default metrics are enabled. With time we will add `encproc` specific metrics for the homomorphic encryption operations.


## Usage

If you want to prototype and experiment with this engine, you can either run it yourself or use an already running instance provided by someone else. The latter is the primary intended use case for this engine. If you are experimenting with a running instance hosted by someone else, you will need to register a data stream with the engine. Feel free to use the [pseudocrypt instance](https://pseudocrypt.site/) The current front-end for registering a stream is simply creating an empty survey. In the near future we will add a separate UI for creating non-survey encrypted streams.

For detailed instructions on how to use this engine, please refer to the documentation and examples provided in this repository. Below, we provide an overview of the API endpoints.

## OpenFHE Functionality

In addition to Lattigo, **encproc** is being extended to support [OpenFHE](https://github.com/openfheorg/openfhe-development), a widely used open-source library for homomorphic encryption. OpenFHE provides a rich set of features and supports multiple HE schemes, including BFV, CKKS, and BGV, enabling more flexible and efficient encrypted computations.

The integration of OpenFHE into encproc is ongoing and experimental. This will allow users to choose between different HE backends according to their requirements and benefit from the latest advancements in the homomorphic encryption community. As with the Lattigo-based functionality, OpenFHE operations are exposed through the same API and client-side WASM modules, making it easy to experiment with both libraries.

> **Note:** The OpenFHE functionality is currently in an early stage and may not be as stable or feature-complete as the Lattigo integration. We welcome feedback and contributions from the community to help improve and expand this support.

## License

This project is licensed under the Apache-2.0 License. See the [LICENSE](LICENSE) file for details.

- [Encproc Homepage](https://pseudocrypt.site/) -- The homepage of this project with a roadmap and an introduction to the overall encproc project.
- Contact: encproc@gmail.com

Happy encrypted processing!
