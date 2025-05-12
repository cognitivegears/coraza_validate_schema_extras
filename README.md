# Coraza Schema Validation Test Project

This project demonstrates and validates the `@validateSchema` operator implementation in [Coraza WAF](https://github.com/corazawaf/coraza), focusing on JSON schema validation capabilities.

## Overview

The Coraza Web Application Firewall (WAF) includes a powerful new `@validateSchema` operator that validates incoming JSON data against formal schemas. This project provides a test server and sample files to verify and demonstrate this functionality.

## Features

- Test server that uses Coraza's schema validation capabilities
- Sample JSON schemas for validation
- Valid and invalid test files for both formats
- Comprehensive configuration examples
- Detailed guides for implementation

## Project Structure

```
├── broken/             # Files with syntax errors
├── invalid/            # Files that don't validate against schemas
├── valid/              # Files that validate successfully against schemas
├── rules/              # ModSecurity rule configurations
│   ├── json_validation.conf
│   ├── xml_validation.conf
│   └── main.conf
├── schemas/            # Schema definition files
│   ├── user.json       # JSON Schema definition
├── server.go           # Test server implementation
├── validate_json_schema_guide.md  # JSON validation guide
```

## How It Works

The project implements a simple HTTP server with Coraza WAF integration that:

1. Listens for POST requests
2. Uses ModSecurity rules to validate the request body against appropriate schemas
3. Returns a success message if validation passes
4. Returns an error if validation fails

### Validation Process

- For JSON content types:
  - Enables JSON body processing
  - Checks for required fields, types, formats, and constraints

## Usage

### Running the Server

```bash
go run server.go
```

By default, the server listens on port 8080. You can specify a different port with the `-port` flag:

```bash
go run server.go -port 9000
```

## Schema Definitions

## Documentation

For detailed information on implementing schema validation in your own applications:

- [JSON Schema Validation Guide](validate_json_schema_guide.md)

These guides provide comprehensive information on configuration requirements, examples, security benefits, and troubleshooting tips.

## Requirements

- Go 1.18 or higher
- Coraza WAF v3
- Docker (for containerized deployment)

## Using a Custom Coraza Branch

This project is configured to use a Git submodule for Coraza, tracking the `feature/schema` branch of the
[`cognitivegears/coraza`](https://github.com/cognitivegears/coraza) repository.

To initialize and update the submodule:
```bash
git submodule update --init
```

After updating submodules, build or run the server as usual:
```bash
go build -o validate-server .
go run server.go
```

## Docker Container Usage

This project includes a `Dockerfile` to build a containerized version of the validation server. This allows for easy testing and deployment, especially when working with custom rule sets.

### Building the Docker Image

Ensure you have initialized the Coraza submodule first:
```bash
git submodule update --init
```

Then, build the Docker image:
```bash
docker build -t coraza-validate-server .
```

### Running the Container

To run the container with the default embedded rules and schemas, exposing port 8080:
```bash
docker run -p 8080:8080 --rm coraza-validate-server
```

The server inside the container will listen on port 8080.

### Running with Custom Rules and Schemas

To test your own Coraza rules and schemas, you can mount a local directory into the container. The container expects the rules directory structure to be:

```
your-custom-rules-dir/
├── rules/
│   └── main.conf  # Your main entry point for rules
│   └── (other .conf files and .json schemas included by main.conf)
```

Mount your local directory (e.g., `./my-custom-rules`) to `/app/rules` inside the container:

```bash
docker run -p 8080:8080 -v $(pwd)/my-custom-rules:/app/rules --rm coraza-validate-server
```

Replace `$(pwd)/my-custom-rules` with the absolute path to your custom rules directory.
The server inside the container will automatically load `main.conf` from the mounted `/app/rules/` directory.

Note: Because of the path used when starting the server, schema definitions must use a relative path starting with `rules/` to be found, i.e. `rules/user.json` instead of `user.json`.

### Testing the Containerized Server

Once the container is running (either with default or custom rules), you can test it using `curl` as described in the [Testing Validation](#testing-validation) section, targeting `http://localhost:8080`.

Example:
```bash
# Test valid JSON against the container
curl -X POST -H "Content-Type: application/json" --data @valid/valid_user.json http://localhost:8080/validate
```

### Viewing Logs

The server logs (including Coraza errors and validation successes/failures) are printed to the standard output of the container. You can view them directly in the terminal where you ran `docker run` or using `docker logs <container_id>` if running detached.

## GitHub Actions: Docker Build and Release

A GitHub Actions workflow is included in `.github/workflows/docker-release.yml`. This workflow automatically builds the Docker image when changes are pushed to the `main` branch. For tagged commits (e.g., `v1.0.0`), it will also create a GitHub Release and attach the built server binary as an asset (Note: It currently does not push the Docker image to a registry).

## License

This project is licensed under the Apache License, Version 2.0 - the same license used by Coraza WAF.
