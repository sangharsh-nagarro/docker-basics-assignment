# Log Management System

This project implements a simple log management system using Go, PostgreSQL, and Redis. It provides an API for inserting and retrieving log entries, with caching capabilities for improved performance.

## Features

- Insert log entries via POST request
- Retrieve logs based on time range and log level
- Redis caching for faster log retrieval
- Dockerized setup for easy deployment

## Tech Stack

- Go (Golang) for the main application
- PostgreSQL for persistent storage
- Redis for caching
- Docker and Docker Compose for containerization

## Setup

### Prerequisites

- Docker
- Docker Compose

### Running the Application

1. Clone the repository:
   ```
   git clone [<repository-url>](https://github.com/sangharsh-nagarro/docker-basics-assignment)
   cd docker-basics-assignment
   ```

2. Build and start the containers:
   ```
   docker-compose up -d
   ```

3. The application should now be running and accessible at `http://localhost:8080`.

## API Endpoints

### POST /api/logs

Insert a new log entry.

**Request Body:**
```json
{
  "log_message": "Your log message",
  "log_level": "info"
}
```

### GET /api/logs

Retrieve logs based on the specified parameters.

**Query Parameters:**
- `since`: (Optional) (e.g., "1h", "24h")
- `level`: (Optional) Log level filter
- `limit`: (Optional) Maximum number of logs to return

**Example:**
```
GET /api/logs?since=1h&level=info&limit=100
```

## Environment Variables

The application uses the following environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string

These are set in the `docker-compose.yml` file.

## Development

To make changes to the project:

1. Modify the Go code as needed.
2. Rebuild the Docker image:
   ```
   docker compose build goapp
   ```
3. Restart the containers:
   ```
   docker compose up -d
   ```

## Troubleshooting

If you encounter any issues:

1. Check the logs of the containers:
   ```
   docker compose logs
   ```
2. Ensure all containers are running:
   ```
   docker compose ps
   ```
3. Verify the environment variables in the `docker-compose.yml` file.

## Contributing

Please read CONTRIBUTING.md for details on our code of conduct, and the process for submitting pull requests.

## License

This project is licensed under the [MIT License](LICENSE).
