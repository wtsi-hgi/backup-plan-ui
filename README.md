# Backup Plan UI

[![CI/CD](https://github.com/wtsi-hgi/backup-plan-ui/actions/workflows/test-and-deploy.yml/badge.svg?branch=main)](https://github.com/sanger/backup-plans-ui/actions/workflows/test-and-deploy.yml)

A web-based user interface for managing backup plans. This application allows users to view, add, edit, and delete backup plan entries stored in a CSV file, SQLite or MySQL database.

## Installation

### Prerequisites

- Go 1.24.4 or later

### Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/sanger/backup-plans-ui.git
   cd backup-plans-ui
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build .
   ```
   
4. Run the application:
   ```bash
   ./backup-plan-ui csv ./data/plan.csv
   ```

5. By default, the application runs on port 4000. You can access it at:
   ```
   http://localhost:4000
   ```

6. To change the port, set the `BACKUP_PLAN_UI_PORT` environment variable:
   ```bash
   export BACKUP_PLAN_UI_PORT=8080
   ./backup-plan-ui csv ./data/plan.csv
   ```

You can also run the app using SQLite backend:
```bash
./backup-plan-ui sqlite ./data/plan.sqlite
```
Or using MySQL backend:
```bash
export MYSQL_HOST=<mysql-db-host>
export MYSQL_PORT=<mysql-db-port>
export MYSQL_USER=<mysql-user>
export MYSQL_PASS=<mysql-password>
export MYSQL_DATABASE=<mysql-db-name>
./backup-plan-ui mysql
```

## Development

### Setting Up Development Environment

Follow installation steps 1-2 to clone the repository and to install dependencies.

### Running the Application Locally

Build and run the application:
   ```bash
   go run main.go csv ./data/plan.csv
   ```

### Running Tests

Run all tests with:
```bash
go test -tags test -v ./...
```
MySQL tests will be skipped unless MySQL variables are set. 

## Deployment

The application is deployed using GitHub Actions.

For production deployment:
1. Merge changes to the `main` branch

For development deployment:
1. Create a Pull Request to the `main` branch

GitHub Actions will automatically build, test, and deploy the application.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Wellcome Sanger Institute - Human Genetics Informatics
