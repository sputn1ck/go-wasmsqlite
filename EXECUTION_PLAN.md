# SQLite WASM Plugin for SQLC - Execution Plan

## Project Overview
Create a SQLC plugin that generates Go and JavaScript code for using SQLite from within the browser via WebAssembly. The plugin itself will be compiled to WASM to run sandboxed within SQLC.

## Architecture

### 1. Plugin Structure
```
sqlc-wasm-plugin/
├── plugin/
│   ├── main.go                 # WASM plugin entry point
│   ├── codegen.go              # Code generation logic
│   ├── templates/
│   │   ├── go_client.tmpl      # Go client template
│   │   ├── js_client.tmpl      # JavaScript client template
│   │   └── types.tmpl          # Shared type definitions
│   └── build.sh                # Build script for WASM
├── runtime/
│   ├── sqlite3.wasm            # SQLite WASM binary
│   ├── sqlite3.mjs             # SQLite JS module
│   ├── bridge.go               # Go-JS bridge code
│   ├── worker.js               # Web Worker for SQLite
│   └── driver.go               # database/sql driver implementation
├── examples/
│   ├── go-example/             # Example Go usage
│   └── js-example/             # Example JS usage
├── sqlc.yaml                   # SQLC configuration
├── schema.sql                  # Database schema
├── queries.sql                 # SQL queries
└── README.md                   # Documentation
```

## Implementation Steps

### Phase 1: Plugin Foundation (Week 1)
1. **Set up plugin project structure**
   - Create plugin directory with main.go
   - Import codegen.pb.go protobuf definitions
   - Set up Go module with dependencies

2. **Implement WASM plugin interface**
   - Parse CodeGenRequest from stdin
   - Implement basic response structure
   - Handle plugin options parsing

3. **Build and test basic plugin**
   - Create build script for WASM compilation
   - Test with simple sqlc configuration
   - Verify plugin loading and execution

### Phase 2: Code Generation Logic (Week 2)
1. **Parse SQLC input**
   - Extract table schemas from catalog
   - Parse query definitions
   - Map SQL types to Go/JS types

2. **Generate Go code**
   - Create struct definitions for tables
   - Generate query methods with proper signatures
   - Implement WASM bridge calls
   - Add context and error handling

3. **Generate JavaScript code**
   - Create TypeScript/JavaScript type definitions
   - Generate async query functions
   - Implement Web Worker communication
   - Add promise-based API

### Phase 3: Runtime Integration (Week 3)
1. **Integrate SQLite WASM runtime**
   - Bundle sqlite3.wasm and sqlite3.mjs
   - Create Web Worker for SQLite operations
   - Implement message passing protocol

2. **Build Go-JS bridge**
   - Implement syscall/js bindings
   - Handle data type conversions
   - Manage connection pooling
   - Add transaction support

3. **Create database/sql driver**
   - Implement driver.Driver interface
   - Support prepared statements
   - Handle result sets and scanning
   - Add connection management

### Phase 4: Testing and Examples (Week 4)
1. **Unit tests**
   - Test code generation logic
   - Verify type mappings
   - Test template rendering

2. **Integration tests**
   - Test Go client with WASM runtime
   - Test JavaScript client in browser
   - Verify cross-platform compatibility

3. **Create examples**
   - Simple CRUD application in Go
   - Web application using generated JS
   - Documentation and usage guides

## Technical Details

### Plugin Implementation
```go
// plugin/main.go
package main

import (
    "encoding/json"
    "io"
    "os"
    
    "github.com/sqlc-dev/sqlc/internal/plugin"
)

func main() {
    // Read CodeGenRequest from stdin
    req := &plugin.CodeGenRequest{}
    data, _ := io.ReadAll(os.Stdin)
    json.Unmarshal(data, req)
    
    // Generate code based on request
    resp := generateCode(req)
    
    // Write CodeGenResponse to stdout
    json.NewEncoder(os.Stdout).Encode(resp)
}
```

### Generated Go Code Example
```go
// generated/db.go
package database

import (
    "context"
    "github.com/project/runtime/wasmsqlite"
)

type Queries struct {
    db *wasmsqlite.DB
}

func (q *Queries) CreateUser(ctx context.Context, username, email string) (*User, error) {
    row := q.db.QueryRowContext(ctx, createUserSQL, username, email)
    var user User
    err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    return &user, err
}
```

### Generated JavaScript Code Example
```javascript
// generated/database.js
export class Database {
    constructor(worker) {
        this.worker = worker;
        this.nextId = 0;
    }
    
    async createUser(username, email) {
        const result = await this.query(
            'INSERT INTO users (username, email) VALUES (?, ?) RETURNING *',
            [username, email]
        );
        return result[0];
    }
    
    async query(sql, params) {
        const id = this.nextId++;
        return new Promise((resolve, reject) => {
            const handler = (e) => {
                if (e.data.id === id) {
                    this.worker.removeEventListener('message', handler);
                    if (e.data.error) {
                        reject(new Error(e.data.error));
                    } else {
                        resolve(e.data.result);
                    }
                }
            };
            this.worker.addEventListener('message', handler);
            this.worker.postMessage({ id, sql, params });
        });
    }
}
```

## Key Considerations

### 1. Type Mapping
- INTEGER → int64 (Go) / number (JS)
- TEXT → string (both)
- BLOB → []byte (Go) / Uint8Array (JS)
- REAL → float64 (Go) / number (JS)
- NULL handling with pointers/null

### 2. WASM Constraints
- No direct filesystem access
- Communication via stdin/stdout only
- Memory limitations
- Binary size optimization

### 3. Performance
- Batch operations support
- Connection pooling
- Prepared statement caching
- Efficient data serialization

### 4. Browser Compatibility
- Web Worker support required
- SharedArrayBuffer for better performance
- CORS headers for WASM loading
- IndexedDB for persistence

## Build Process

### Plugin Build
```bash
# Build the WASM plugin
GOOS=wasip1 GOARCH=wasm go build -o sqlc-wasm-sqlite.wasm ./plugin

# Calculate SHA256 for sqlc.yaml
sha256sum sqlc-wasm-sqlite.wasm
```

### Runtime Build
```bash
# Build Go WASM runtime
GOOS=js GOARCH=wasm go build -o runtime.wasm ./runtime

# Bundle JavaScript files
npm run build
```

## Testing Strategy

1. **Plugin Tests**
   - Mock CodeGenRequest inputs
   - Verify generated code compiles
   - Test template rendering

2. **Runtime Tests**
   - Test in Node.js with WASM support
   - Browser testing with Playwright
   - Cross-browser compatibility

3. **Integration Tests**
   - End-to-end query execution
   - Transaction handling
   - Error scenarios

## Dependencies

### Go Dependencies
- `google.golang.org/protobuf` - Protocol buffers
- `github.com/sqlc-dev/sqlc` - SQLC types (codegen.pb.go)
- Standard library only for WASM build

### JavaScript Dependencies
- SQLite WASM official build
- No external runtime dependencies
- TypeScript for type definitions (optional)

## Success Criteria

1. Plugin successfully generates both Go and JS code
2. Generated code compiles without errors
3. Queries execute correctly in browser environment
4. Performance acceptable for typical web applications
5. Documentation and examples are clear and complete

## Timeline

- **Week 1**: Plugin foundation and basic code generation
- **Week 2**: Complete code generation for all query types
- **Week 3**: Runtime integration and testing
- **Week 4**: Documentation, examples, and optimization

## Next Steps

1. Review and approve this execution plan
2. Set up the plugin project structure
3. Implement the basic WASM plugin interface
4. Begin code generation implementation
5. Integrate with SQLite WASM runtime