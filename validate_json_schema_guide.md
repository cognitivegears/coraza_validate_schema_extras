# ModSecurity @validateSchema Operator Guide for JSON

## Overview

The `@validateSchema` operator in ModSecurity is a powerful tool for validating XML and JSON data against formal schemas. This document focuses on JSON schema validation, providing comprehensive information on how to implement and configure this security feature.

## How @validateSchema Works

The `@validateSchema` operator validates incoming JSON content against a JSON Schema file. It returns `true` when validation fails (meaning the input doesn't conform to the expected schema or syntax), allowing rules to trigger defensive actions.

Key features:
- Validates the parsed JSON structure from the request body processor
- Executes during phase 2 (request body processing)
- Requires explicit JSON processing to be enabled
- Provides comprehensive structure, type, and constraint validation

## Configuration Requirements

### 1. Enable JSON Body Processing

Before schema validation can occur, ModSecurity must parse the incoming request as JSON:

```apache
# Enable JSON processing for JSON content types
SecRule REQUEST_HEADERS:Content-Type "application/json" \
  "id:1000,phase:1,pass,t:lowercase,ctl:requestBodyProcessor=JSON"
```

### 2. Schema Validation Rule

The basic syntax for JSON schema validation:

```apache
SecRule REQUEST_BODY "@validateSchema /path/to/schema.json" \
  "id:1001,phase:2,deny,status:400,log,msg:'JSON schema validation failed'"
```

### 3. Error Handling

It's crucial to handle JSON parsing errors to prevent bypass attacks:

```apache
SecRule REQBODY_PROCESSOR_ERROR "!@eq 0" \
  "id:1002,phase:2,deny,status:400,log,msg:'JSON parsing error: %{REQBODY_PROCESSOR_ERROR_MSG}'"
```

## JSON Schema Validation Examples

### Example 1: Basic API Validation

```apache
# Enable JSON processing for JSON requests
SecRule REQUEST_HEADERS:Content-Type "application/json" \
  "id:2000,phase:1,pass,t:lowercase,ctl:requestBodyProcessor=JSON"

# Validate against API schema
SecRule REQUEST_BODY "@validateSchema /opt/modsecurity/schemas/api-schema.json" \
  "id:2001,phase:2,deny,status:400,log,auditlog,msg:'Invalid JSON structure'"

# Handle JSON parsing errors
SecRule REQBODY_PROCESSOR_ERROR "!@eq 0" \
  "id:2002,phase:2,deny,status:400,log,msg:'JSON processing error: %{REQBODY_PROCESSOR_ERROR_MSG}'"
```

### Example 2: JSON Schema with Required Fields

Example JSON schema (`user-schema.json`):

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "email", "age"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 2,
      "maxLength": 100
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "minimum": 18,
      "maximum": 120
    },
    "role": {
      "type": "string",
      "enum": ["admin", "user", "guest"]
    },
    "settings": {
      "type": "object",
      "properties": {
        "notifications": {
          "type": "boolean"
        },
        "theme": {
          "type": "string"
        }
      }
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "additionalProperties": false
}
```

ModSecurity rule to validate against this schema:

```apache
SecRule REQUEST_BODY "@validateSchema /path/to/user-schema.json" \
  "id:4000,phase:2,deny,status:400,log,msg:'Invalid user JSON format'"
```

This validation will check:
1. That the JSON syntax is valid
2. That required properties (name, email, age) exist with correct types
3. That string lengths are within specified ranges
4. That numeric values are within acceptable ranges
5. That "role" contains only allowed values (admin, user, guest)
6. That no additional properties exist beyond those defined
7. That nested objects and arrays conform to their specified structures

## Advanced Use Cases

### Chaining with Other Security Checks

```apache
SecRule REQUEST_BODY "@validateSchema /schemas/payment-schema.json" \
  "chain,id:5000,phase:2,deny,status:400,log,msg:'Payment validation failed'"
SecRule JSON:.creditCard.number "@validateLuhn"
```

### Implementing Positive Security Model

JSON schema validation can form the foundation of a positive security model:

```apache
# Default deny
SecDefaultAction "phase:2,deny,status:400,log,auditlog"

# Enable JSON processing
SecRule REQUEST_HEADERS:Content-Type "application/json" \
  "id:6000,phase:1,pass,t:lowercase,ctl:requestBodyProcessor=JSON"

# Block on JSON parsing errors
SecRule REQBODY_PROCESSOR_ERROR "!@eq 0" \
  "id:6001,phase:2,log,msg:'JSON processing error'"

# Allow only if schema validates (note 'chain' with '@eq false')
SecRule REQUEST_BODY "@validateSchema /schemas/api-schema.json" \
  "chain,id:6002,phase:2"
SecRule &TX:blocked "@eq 0" "pass"

# All other requests denied by default
```

## Performance Considerations

- Schema validation is resource-intensive
- The operator uses lazy initialization to avoid schema processing at startup
- Parsed schemas are cached after first use to improve performance
- Consider pre-loading schemas at server startup if possible
- Test validation performance with production-like traffic patterns
- For large JSON payloads, consider validating only specific paths or subschemas

## Implementation Details

The implementation uses established JSON Schema validation libraries:
- Uses a full JSON Schema validator (github.com/kaptinlin/jsonschema)
- Supports JSON Schema draft specifications
- Performs comprehensive validation of all aspects of the JSON Schema including:
  - Type validation
  - Format validation (email, dates, etc.)
  - Numeric constraints (min, max)
  - String constraints (min/max length)
  - Array validation
  - Nested object validation
  - Required fields

## Security Benefits

JSON schema validation provides several security benefits:
- Prevents JSON parameter tampering and injection
- Mitigates JSON DoS attacks
- Blocks malformed JSON exploitation
- Enforces proper data structure and typing
- Rejects unexpected content structures
- Reduces attack surface by validating before business logic processing
- Helps mitigate prototype pollution attacks (with appropriate settings)
- Ensures data integrity before processing

## Troubleshooting

Common issues and solutions:

1. **Rule not triggering**: Ensure JSON body processor is enabled in phase 1
2. **False positives**: Validate schema against known-good requests
3. **Performance issues**: Consider validating specific JSON paths instead of entire document
4. **Schema not found**: Verify file paths and permissions
5. **Format validators failing**: Ensure the JSON Schema library supports the formats you're using

Debugging techniques:

```apache
# Log validation attempts without blocking
SecRule REQUEST_BODY "@validateSchema /path/to/schema.json" \
  "id:7000,phase:2,pass,log,auditlog,msg:'Schema validation result'"

# Log specific JSON properties for debugging
SecRule JSON:.email "@rx .*" \
  "id:7001,phase:2,pass,log,msg:'Email value: %{MATCHED_VAR}'"

# Detailed debug logging
SecDebugLog /var/log/modsec_debug.log
SecDebugLogLevel 9
```

## Conclusion

The `@validateSchema` operator provides robust JSON validation capabilities in ModSecurity, enabling comprehensive structural and content validation against formal JSON Schema definitions. When properly implemented with appropriate error handling, it forms a powerful defense against JSON-based attacks and malformed data.
