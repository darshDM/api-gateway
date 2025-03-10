# API Gateway

## Overview
This API Gateway is built using Golang and provides essential functionalities such as request ID generation and rate limiting. It reads a list of services from a `config.yaml` file and forwards incoming requests to the appropriate backend service based on URL prefixes.

## Features
- **Request ID Generation**: Each incoming request is assigned a unique request ID for tracking.
- **Rate Limiting**: Limits the number of requests per second using `tollbooth`.
- **Service Routing**: Routes requests based on URL prefixes defined in `config.yaml`.
- **Logging**: Uses `logrus` for structured logging.
- **Authentication**: Basic Authentication using `x-api-key` by matching same in `config.yaml`
- **Load balancer**: Round robin technique for load balancing

## Todo
- [x] Service based rate limit
- [x] Refine Path handling
- [x] Load balancing
- [ ] Circuit breaking
- [ ] Context checking

## Configuration

The services to be proxied are defined in `config.yaml`. Example:

```yaml
servers:
  - name: "product"
    hosts: 
      - "http://product.api:3000"
    prefix: "/product"
    port: 3000
    api_key: "uit39t8tvjwrfj05ldsf"
    rate_limit: 5
    
  - name: "vendor"
    hosts: 
      - "http://vendor.api:3001"
    prefix: "/vendor"
    port: 3001
    api_key: "sdfjfhasdjdf123ksmdf"
    rate_limit: 2
```

## Running the API Gateway

### Prerequisites
- Golang installed
- Two test HTTP servers running on different ports
- `/etc/hosts` configured to map service names to local IP

### Steps

1. Clone the repository:
   ```sh
   git clone <repo-url>
   cd api-gateway
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Start the API Gateway:
   ```sh
   go run main.go
   ```
4. The gateway runs on port `8001` by default.

## Testing the API Gateway

### Running Mock Services
You can start two mock HTTP servers to simulate backend services:

```sh
# Start product service and vendor service by navigating to respective folders
go run main.go

```

### Making Requests
Once the gateway is running, you can test requests:

```sh
curl -v -H "x-api-key:sdfjfhasdjdf123ksmdf" "localhost:8001/vendor/new"
curl -v -H "x-api-key:uit39t8tvjwrfj05ldsf" "localhost:8001/product/test"
```

### Rate Limiting Test
By default, the rate limit is set to **2 requests per second**. Exceeding this will result in a `429 Too Many Requests` response.

## Logs
The gateway logs incoming requests, service calls, and errors. Example log output:

```json
{
  "service": "api-gateway",
  "requestId": "abc123",
  "IP": "127.0.0.1",
  "Method": "GET",
  "URL": "/product",
  "message": "New Request"
}
```

## Contributing
Feel free to fork the repository and submit pull requests.

## License
This project is licensed under the MIT License.