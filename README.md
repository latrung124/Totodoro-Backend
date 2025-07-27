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