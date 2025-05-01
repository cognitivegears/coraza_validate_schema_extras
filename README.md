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

1. Listens for POST requests on the `/validate` endpoint
2. Uses ModSecurity rules to validate the request body against appropriate schemas
3. Returns a success message if validation passes
4. Returns an error if validation fails

### Validation Process

- For JSON content types:
  - Enables JSON body processing
  - Validates against the `schemas/user.json` schema
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

### Testing Validation

Test successful validation with valid files:

```bash
# Test valid JSON
curl -X POST -H "Content-Type: application/json" --data @valid/valid_user.json http://localhost:8080/validate

# Test valid XML
curl -X POST -H "Content-Type: application/xml" --data @valid/valid_user.xml http://localhost:8080/validate
```

Test failed validation with invalid files:

```bash
# Test invalid JSON
curl -X POST -H "Content-Type: application/json" --data @invalid/invalid_user.json http://localhost:8080/validate

# Test invalid XML
curl -X POST -H "Content-Type: application/xml" --data @invalid/invalid_user.xml http://localhost:8080/validate
```

## Schema Definitions

### JSON Schema

The JSON schema (`schemas/user.json`) validates user objects with:
- Required fields: name, email, age
- Optional fields: initials, role, settings
- Type validation for all fields
- Format validation for email
- Pattern validation for initials
- Range constraints for numeric values
- Enumeration restrictions for specific fields
- Nested object validation

## Implementation Details

The ModSecurity rules in the `rules/` directory show how to:
- Enable the appropriate body processor based on content type
- Handle parsing errors to prevent bypass attacks
- Configure schema validation with the `@validateSchema` operator

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
│   └── (other .conf files included by main.conf)
└── schemas/
    └── (your .json schema files)
```

Mount your local directory (e.g., `./my-custom-rules`) to `/etc/coraza/rules` inside the container:

```bash
docker run -p 8080:8080 -v $(pwd)/my-custom-rules:/etc/coraza/rules --rm coraza-validate-server
```

Replace `$(pwd)/my-custom-rules` with the absolute path to your custom rules directory.
The server inside the container will automatically load `main.conf` from the mounted `/etc/coraza/rules/rules/` directory and serve schemas from `/etc/coraza/rules/schemas/`.

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
