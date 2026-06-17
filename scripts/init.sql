  CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
  );


  CREATE TABLE IF NOT EXISTS polls (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    question TEXT NOT NULL,
    code VARCHAR(4) DEFAULT NULL UNIQUE,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  
  );

  CREATE TABLE IF NOT EXISTS options (
    id SERIAL PRIMARY KEY,
    poll_id INTEGER REFERENCES  polls(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    votes_count INTEGER DEFAULT 0
  );

  CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    option_id INTEGER NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS api_logs (
      id SERIAL PRIMARY KEY,
      user_id INTEGER,         
      action TEXT,             
      status_code INTEGER,      
      path TEXT,                
      latency_ms INTEGER,       
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );