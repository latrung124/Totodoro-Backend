# Totodoro-Backend

Microservices Monolith for Totodoro-Backend.

## Setup Instructions

Follow these steps to set up the project on your local machine:

### Prerequisites

Ensure you have the following installed on your system:
- [Go](https://go.dev/dl/) (version 1.20 or later)
- [Protocol Buffers Compiler (protoc)](https://github.com/protocolbuffers/protobuf/releases) (version 32.0-rc1 or later)
- Git
- A PostgreSQL database (or any database supported by the project)

### Clone the Repository

```bash
git clone https://github.com/latrung124/Totodoro-Backend.git
cd Totodoro-Backend
```

Install Dependencies
Run the following command to install Go dependencies:

```bash
go mod tidy
```

### Setting Up Protocol Buffers
1. Download and Install protoc:
- If protoc is not already installed, the setup script will download and extract it automatically.

2. Install Go Plugins for protoc:
- Run the following commands to install the required plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
- Ensure $GOBIN is in your system's PATH.

3. Generate .pb.go Files:

- Run the setup script to generate the .pb.go files:
```bash
./scripts/setup-windows.bat
```

### Database Configuration
1. Create the required databases for the project (e.g., UserDB, PomodoroDB, etc.).
2. Update the database connection strings in the environment variables or configuration files.

### Build the Project
Run the following command to build the project:
```bash
./scripts/build-windows.bat
```

### Run the Project
After building, you can run the project using:

```bash
./bin/totodoro-backend.exe
```

### Running Unit Tests

To ensure the project is functioning as expected, you can run the unit tests provided in the repository.

1. **Set Up the Test Environment**:
   - Ensure the `.env` file is properly configured for testing. Use environment variables like `TEST_USER_DB_URL` to point to a test database.

2. **Run All Tests**:
   - Use the following command to run all unit tests in the project:
     ```bash
     go test ./...
     ```

3. **Run a Specific Test**:
   - To run a specific test, use the `-run` flag with the test name:
     ```bash
     go test -run TestCreateUser ./...
     ```

4. **Verbose Output**:
   - Use the `-v` flag to see detailed output for each test:
     ```bash
     go test -v ./...
     ```

5. **Check Coverage**:
   - To check test coverage, use the `-cover` flag:
     ```bash
     go test -cover ./...
     ```

6. **Debugging Tests**:
   - Add `t.Log` or `log.Printf` statements in your test files to debug failing tests. Use the `-v` flag to see the logs.