# Coraza Schema Validation Test Project

This project demonstrates and validates the `@validateSchema` operator implementation in [Coraza WAF](https://github.com/corazawaf/coraza), focusing on both JSON and XML schema validation capabilities.

## Overview

The Coraza Web Application Firewall (WAF) includes a powerful new `@validateSchema` operator that validates incoming JSON and XML data against formal schemas. This project provides a test server and sample files to verify and demonstrate this functionality.

## Features

- Test server that uses Coraza's schema validation capabilities
- Sample JSON and XML schemas for validation
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
│   └── user.xsd        # XML Schema Definition
├── server.go           # Test server implementation
├── validate_json_schema_guide.md  # JSON validation guide
└── validate_xml_schema_guide.md   # XML validation guide
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

- For XML content types:
  - Enables XML body processing
  - Validates against the `schemas/user.xsd` schema
  - Verifies elements, attributes, types, and constraints

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

### XML Schema

The XML schema (`schemas/user.xsd`) validates user elements with:
- Required attribute: id
- Required elements: name, email, age
- Optional elements: role
- Type validation for all elements
- Enumeration restrictions for the role element

## Implementation Details

The ModSecurity rules in the `rules/` directory show how to:
- Enable the appropriate body processor based on content type
- Handle parsing errors to prevent bypass attacks
- Configure schema validation with the `@validateSchema` operator

## Documentation

For detailed information on implementing schema validation in your own applications:

- [JSON Schema Validation Guide](validate_json_schema_guide.md)
- [XML Schema Validation Guide](validate_xml_schema_guide.md)

These guides provide comprehensive information on configuration requirements, examples, security benefits, and troubleshooting tips.

## Requirements

- Go 1.18 or higher
- Coraza WAF v3

## License

This project is licensed under the Apache License, Version 2.0 - the same license used by Coraza WAF.
