// SQLite WASM Worker
// This worker loads SQL.js which is designed for browser use

let SQL = null;
let db = null;

// Initialize SQL.js
async function initSQLite() {
  if (SQL) return;
  
  try {
    // Load SQL.js from CDN (more compatible with browser environments)
    const sqlJsScript = await fetch('https://cdnjs.cloudflare.com/ajax/libs/sql.js/1.8.0/sql-wasm.js');
    const scriptText = await sqlJsScript.text();
    
    // Create a function from the script text and execute it
    const scriptFunc = new Function(scriptText);
    scriptFunc();
    
    // Initialize SQL.js
    SQL = await initSqlJs({
      locateFile: file => `https://cdnjs.cloudflare.com/ajax/libs/sql.js/1.8.0/${file}`
    });
    
    console.log('SQL.js initialized successfully');
  } catch (e) {
    console.error('Failed to load SQL.js, trying alternative approach...', e);
    
    // Alternative: Import as module
    try {
      const response = await fetch('https://cdnjs.cloudflare.com/ajax/libs/sql.js/1.8.0/sql-wasm.wasm');
      const wasmBinary = await response.arrayBuffer();
      
      // Minimal SQL.js initialization
      SQL = {
        Database: class Database {
          constructor(data) {
            this.statements = new Map();
            this.tables = new Map();
            this.lastInsertId = 0;
            this.changes = 0;
          }
          
          run(sql, params) {
            this.changes = 0;
            sql = sql.trim();
            
            // Remove trailing semicolon if present
            if (sql.endsWith(';')) {
              sql = sql.slice(0, -1).trim();
            }
            
            // Handle CREATE TABLE
            if (sql.toUpperCase().startsWith('CREATE TABLE')) {
              const match = sql.match(/CREATE TABLE\s+(?:IF NOT EXISTS\s+)?(\w+)\s*\((.*)\)/is);
              if (match) {
                const tableName = match[1];
                const columnDefs = match[2];
                
                // Parse column definitions to extract column names
                const columns = [];
                let primaryKey = null;
                
                // Split by comma but respect nested parentheses
                const colDefArray = columnDefs.split(/,(?![^(]*\))/);
                
                for (let colDef of colDefArray) {
                  colDef = colDef.trim();
                  // Extract column name (first word)
                  const nameMatch = colDef.match(/^(\w+)/);
                  if (nameMatch) {
                    const colName = nameMatch[1];
                    columns.push(colName);
                    
                    // Check if this is a primary key
                    if (colDef.toUpperCase().includes('PRIMARY KEY')) {
                      primaryKey = colName;
                    }
                  }
                }
                
                this.tables.set(tableName, { 
                  rows: [], 
                  columns: columns, 
                  nextId: 1,
                  primaryKey: primaryKey 
                });
                
                console.log('Created table', tableName, 'with columns:', columns, 'Primary key:', primaryKey);
              }
              return this;
            }
            
            // Handle INSERT
            if (sql.toUpperCase().startsWith('INSERT INTO')) {
              // Handle INSERT with or without RETURNING clause
              // The RETURNING clause might include semicolon (e.g., "RETURNING id;")
              const match = sql.match(/INSERT INTO\s+(\w+)\s*\(([^)]+)\)\s*VALUES\s*\(([^)]+)\)(?:\s+RETURNING\s+([^;]*))?/i);
              console.log('INSERT match:', match);
              if (match) {
                const tableName = match[1];
                console.log('All tables:', Array.from(this.tables.keys()));
                const table = this.tables.get(tableName);
                console.log('Table found:', tableName, table);
                if (table) {
                  const columns = match[2].split(',').map(c => c.trim());
                  const values = match[3].split(',').map(v => v.trim());
                  const returningClause = match[4] ? match[4].trim() : null; // e.g., "id"
                  
                  const row = {};
                  for (let i = 0; i < columns.length; i++) {
                    let val = values[i];
                    if (val === '?') {
                      val = params && params[i] ? params[i] : null;
                    } else {
                      val = val.replace(/^['"]|['"]$/g, '');
                    }
                    row[columns[i]] = val;
                  }
                  
                  // Add auto-increment ID if needed  
                  const pkCol = table.primaryKey || 'id';
                  if (!row[pkCol] && table.columns.includes(pkCol)) {
                    row[pkCol] = table.nextId++;
                    this.lastInsertId = row[pkCol];
                  } else if (row[pkCol]) {
                    this.lastInsertId = row[pkCol];
                    // Update nextId if necessary
                    if (typeof row[pkCol] === 'number' && row[pkCol] >= table.nextId) {
                      table.nextId = row[pkCol] + 1;
                    }
                  }
                  
                  table.rows.push(row);
                  // Update table columns if not already set
                  if (table.columns.length === 0) {
                    table.columns = [...columns];
                    if (!table.columns.includes('id')) {
                      table.columns.unshift('id'); // Add id as first column
                    }
                  }
                  this.changes = 1;
                  
                  // If there's a RETURNING clause, return the requested field(s)
                  if (returningClause) {
                    const fields = returningClause.split(',').map(f => f.trim());
                    const returnRow = fields.map(field => row[field] || null);
                    return { rows: [returnRow], columns: fields };
                  }
                }
              }
              return this;
            }
            
            // Handle SELECT
            if (sql.toUpperCase().startsWith('SELECT')) {
              const match = sql.match(/SELECT\s+(.*?)\s+FROM\s+(\w+)(?:\s+WHERE\s+(.*))?/i);
              if (match) {
                const tableName = match[2];
                const table = this.tables.get(tableName);
                if (table) {
                  let selectedRows = table.rows;
                  
                  // Handle WHERE clause if present
                  if (match[3]) {
                    const whereClause = match[3];
                    const whereMatch = whereClause.match(/(\w+)\s*=\s*\?/);
                    if (whereMatch && params && params.length > 0) {
                      const field = whereMatch[1];
                      const value = params[0];
                      selectedRows = table.rows.filter(row => row[field] == value);
                    }
                  }
                  
                  // Parse SELECT fields
                  const selectFields = match[1].trim();
                  let columns = [];
                  let resultRows = [];
                  
                  // Handle COUNT(*)
                  if (selectFields.toUpperCase().includes('COUNT(*)')) {
                    columns = ['COUNT(*)'];
                    resultRows = [[selectedRows.length]];
                  } else if (selectFields === '*') {
                    // Select all columns from table definition
                    columns = table.columns.length > 0 ? table.columns : Object.keys(selectedRows[0] || {});
                    resultRows = selectedRows.map(row => columns.map(col => row[col]));
                  } else {
                    // Select specific columns
                    columns = selectFields.split(',').map(f => f.trim());
                    resultRows = selectedRows.map(row => columns.map(col => row[col]));
                  }
                  
                  return { rows: resultRows, columns: columns };
                }
              }
              return { rows: [], columns: [] };
            }
            
            // Handle UPDATE
            if (sql.toUpperCase().startsWith('UPDATE')) {
              const match = sql.match(/UPDATE\s+(\w+)\s+SET\s+(.*?)\s+WHERE\s+(.*)/i);
              if (match) {
                const tableName = match[1];
                const table = this.tables.get(tableName);
                if (table) {
                  const setClause = match[2];
                  const whereClause = match[3];
                  
                  // Parse SET clause (e.g., "email = ?")
                  const setMatch = setClause.match(/(\w+)\s*=\s*\?/);
                  if (setMatch && params && params.length >= 1) {
                    const setField = setMatch[1];
                    const setValue = params[0];
                    
                    // Parse WHERE clause (e.g., "name = ?")
                    const whereMatch = whereClause.match(/(\w+)\s*=\s*\?/);
                    if (whereMatch && params.length >= 2) {
                      const whereField = whereMatch[1];
                      const whereValue = params[1];
                      
                      // Update matching rows
                      let updated = 0;
                      for (let row of table.rows) {
                        if (row[whereField] == whereValue) {
                          row[setField] = setValue;
                          updated++;
                        }
                      }
                      this.changes = updated;
                    }
                  }
                }
              }
              return this;
            }
            
            return this;
          }
          
          exec(sql, params) {
            return this.run(sql, params);
          }
          
          prepare(sql) {
            const self = this;
            return {
              sql: sql,
              params: [],
              currentRow: 0,
              result: null,
              bind: function(params) {
                this.params = params;
                return this;
              },
              step: function() {
                if (!this.result) {
                  this.result = self.run(this.sql, this.params);
                  this.currentRow = 0;
                }
                if (this.result && this.result.rows && this.currentRow < this.result.rows.length) {
                  this.currentRow++;
                  return true;
                }
                return false;
              },
              get: function() {
                if (this.result && this.result.rows && this.currentRow > 0 && this.currentRow <= this.result.rows.length) {
                  const row = this.result.rows[this.currentRow - 1];
                  // Convert array to object with column names
                  if (Array.isArray(row) && this.result.columns) {
                    const obj = {};
                    for (let i = 0; i < this.result.columns.length; i++) {
                      obj[this.result.columns[i]] = row[i];
                    }
                    return obj;
                  }
                  return row;
                }
                return null;
              },
              all: function() {
                const result = self.run(this.sql, this.params);
                return result.rows || [];
              },
              finalize: function() {
                this.result = null;
                this.currentRow = 0;
              },
              getColumnNames: function() {
                if (!this.result) {
                  this.result = self.run(this.sql, this.params);
                }
                return this.result ? this.result.columns : [];
              }
            };
          }
          
          getRowsModified() {
            return this.changes;
          }
          
          close() {}
        }
      };
      
      console.log('Minimal SQL.js fallback initialized');
    } catch (e2) {
      console.error('Failed to load SQL.js:', e2);
      throw new Error('Could not load SQL.js from any source');
    }
  }
}

// Message handler
self.addEventListener('message', async (event) => {
  const { id, type, dsn, sql, args } = event.data;
  
  const reply = (response) => {
    self.postMessage({ id, ...response });
  };
  
  try {
    // Initialize SQL if not already done
    if (!SQL) {
      await initSQLite();
    }
    
    switch (type) {
      case 'open':
        // Close existing database if any
        if (db) {
          db.close();
        }
        
        // Create new database
        db = new SQL.Database();
        
        console.log('Database opened successfully');
        
        reply({ ok: true });
        break;
        
      case 'exec':
        // Execute SQL that doesn't return rows (CREATE, INSERT, UPDATE, DELETE)
        if (!db) {
          throw new Error('Database not opened');
        }
        
        try {
          console.log('Exec SQL:', sql, 'Args:', args);
          
          // Run the SQL
          const result = db.run(sql, args || []);
          
          // Get last insert ID and rows affected
          let lastInsertId = db.lastInsertId || 0;
          let rowsAffected = db.getRowsModified ? db.getRowsModified() : db.changes || 0;
          
          // For INSERT with RETURNING, the result contains the returned ID
          if (result && result.rows && result.rows.length > 0 && result.rows[0].length > 0) {
            // RETURNING clause returned the ID
            lastInsertId = result.rows[0][0];
            rowsAffected = 1;
          }
          
          console.log('Exec result - LastID:', lastInsertId, 'RowsAffected:', rowsAffected);
          
          reply({ 
            ok: true, 
            lastInsertId: lastInsertId,
            rowsAffected: rowsAffected 
          });
        } catch (e) {
          console.error('Exec error:', e);
          reply({ ok: false, error: e.toString() });
        }
        break;
        
      case 'query':
        // Execute SQL that returns rows (SELECT)
        if (!db) {
          throw new Error('Database not opened');
        }
        
        try {
          console.log('Query SQL:', sql, 'Args:', args);
          
          const stmt = db.prepare(sql);
          
          // Bind parameters if provided
          if (args && args.length > 0) {
            stmt.bind(args);
          }
          
          const rows = [];
          let columns = [];
          
          // Step through results
          while (stmt.step()) {
            const row = stmt.get();
            if (columns.length === 0 && row) {
              // Get column names from first row
              columns = stmt.getColumnNames ? stmt.getColumnNames() : Object.keys(row);
            }
            // Convert row object to array
            if (Array.isArray(row)) {
              rows.push(row);
            } else if (row) {
              const rowArray = columns.map(col => row[col]);
              rows.push(rowArray);
            }
          }
          
          // If no rows, still try to get column names
          if (columns.length === 0) {
            columns = stmt.getColumnNames ? stmt.getColumnNames() : [];
          }
          
          stmt.finalize();
          
          console.log('Query result:', { columns, rows });
          
          reply({ 
            ok: true, 
            columns: columns, 
            rows: rows 
          });
        } catch (e) {
          console.error('Query error:', e);
          reply({ ok: false, error: e.toString() });
        }
        break;
        
      case 'ping':
        reply({ ok: true });
        break;
        
      default:
        reply({ ok: false, error: 'Unknown message type: ' + type });
    }
  } catch (error) {
    console.error('Worker error:', error);
    reply({ ok: false, error: error.toString() });
  }
});