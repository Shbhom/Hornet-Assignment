# Product Service Microservice
A RESTful microservice for managing products with PostgreSQL and Redis caching written in GO.

## Setup

### Environment Variable
create .env file with the following configuration

```
    DB_URL="postgres://postgres:postgres@postgres/products?sslmode=disable"
    REDIS_HOST="redis"
    REDIS_PORT="6379"
    REDIS_PASS=""
    REDIS_DB="0"
```

## Running the application

### building the application

Use the following command to build the application 
```bash
docker compose build
```

### Run the application

Use the following command to run the application
```bash
docker compose up -d
```

## Application logs

to view the application logs use the following command
```bash
docker compose logs app -f
```

## API Endpoints
1. Get All Products
    
    Request GET /products
    ```bash
        curl http://localhost:8080/products?page=2&limit=5
    ```
    Response:
    ```json
    {
      "data": [
        {"id":6,"name":"Mouse","price":29.99},
        {"id":7,"name":"Tablet","price":299.99}
      ],
      "metadata": {
        "currentPage": 2,
        "totalProducts": 10,
        "limit": 5,
        "totalpages": 2
      }
    }
2. Get Product by ID

    GET /products/:id
    ```bash
    curl http://localhost:8080/products/3
    ```
    Response:
    ```json
    {"id":3,"name":"Headphones","price":149.99}
    ```
3. Create Product

    POST /products
    ```bash
    curl -X POST http://localhost:8080/products \
      -H "Content-Type: application/json" \
      -d '{"name": "Gaming Console", "Price": 499.99}'
    ```
    Response:
    ```json
    {"name":"Gaming Console","Price":499.99}
    ```

4. Update Product

    PUT /products/:id

    ```bash
    curl -X PUT http://localhost:8080/products/5 \
      -H "Content-Type: application/json" \
      -d '{"name": "4K Monitor", "Price": 299.99}'
    ```
    Response:
    ```json
    {"id":5,"name":"4K Monitor","price":299.99}
    ```

5. Delete Product

    DELETE /products/:id
    ```bash

    curl -X DELETE http://localhost:8080/products/8
    ```

    Response: 204 No Content