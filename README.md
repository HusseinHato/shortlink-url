
# 🔗 Shortlink URL Service

  

A high-performance URL shortener service built with Go, using Base62 encoding and PostgreSQL for persistent storage. This service converts long URLs into short, shareable links with automatic redirection.

  

## 📋 Table of Contents

  

-  [Features](#-features)

-  [Architecture](#-architecture)

-  [Technology Stack](#-technology-stack)

-  [Prerequisites](#-prerequisites)

-  [Installation](#-installation)

-  [Configuration](#-configuration)

-  [Running the Service](#-running-the-service)

-  [API Documentation](#-api-documentation)

-  [Usage Examples](#-usage-examples)

-  [Database Schema](#-database-schema)

-  [How It Works](#-how-it-works)

-  [Development](#-development)

-  [Troubleshooting](#-troubleshooting)

-  [License](#-license)

  

## ✨ Features

  

-  **Base62 Encoding**: Generates short, URL-safe codes using alphanumeric characters (0-9, a-z, A-Z)

-  **PostgreSQL Backend**: Reliable persistent storage with automatic schema initialization

-  **Fast Redirects**: 301 permanent redirects for optimal SEO and caching

-  **RESTful API**: Clean JSON API for creating and managing short URLs

-  **Health Checks**: Built-in health endpoint for monitoring and load balancers

-  **CORS Support**: Cross-origin resource sharing enabled for web applications

-  **Auto-incrementing IDs**: Sequential ID generation ensures unique short codes

-  **Stats Endpoint**: Retrieve information about shortened URLs

-  **Logging**: Request logging and error tracking built-in

  

## 🏗️ Architecture

  

The service uses a simple but effective architecture:

  

```

┌─────────────┐

│ Client │

└──────┬──────┘

│

▼

┌─────────────────────────────────┐

│ Echo Web Framework │

│ ┌──────────────────────────┐ │

│ │ POST /shorten │ │

│ │ GET /:shortCode │ │

│ │ GET /api/stats/:code │ │

│ │ GET /health │ │

│ └──────────────────────────┘ │

└────────────┬────────────────────┘

│

▼

┌─────────────────────────────────┐

│ Base62 Encoder │

│ (ID → Short Code Conversion) │

└────────────┬────────────────────┘

│

▼

┌─────────────────────────────────┐

│ PostgreSQL Database │

│ ┌───────────────────────────┐ │

│ │ urls table │ │

│ │ - id (SERIAL) │ │

│ │ - short_code (UNIQUE) │ │

│ │ - original_url │ │

│ │ - created_at │ │

│ └───────────────────────────┘ │

└─────────────────────────────────┘

```

  

## 🛠️ Technology Stack

  

-  **Language**: Go 1.25.4

-  **Web Framework**: [Echo v4](https://echo.labstack.com/) - High performance, minimalist Go web framework

-  **Database**: PostgreSQL with `lib/pq` driver

-  **Encoding**: Custom Base62 implementation

-  **Dependencies**:

-  `github.com/labstack/echo/v4` - Web framework

-  `github.com/lib/pq` - PostgreSQL driver

  

## 📦 Prerequisites

  

Before running this service, ensure you have:

  

-  **Go**: Version 1.20 or higher ([Download](https://golang.org/dl/))

-  **PostgreSQL**: Version 12 or higher ([Download](https://www.postgresql.org/download/))

-  **Git**: For cloning the repository

  

## 🚀 Installation

  

### 1. Clone the Repository

  

```bash

git  clone  <repository-url>

cd  shortlink-url

```

  

### 2. Install Dependencies

  

```bash

go  mod  download

```

  

### 3. Set Up PostgreSQL Database

  

Create a new database for the URL shortener:

  

```bash

# Connect to PostgreSQL

psql  -U  postgres

  

# Create database

CREATE  DATABASE  urlshortener;

  

# Exit psql

\q

```

  

## ⚙️ Configuration

  

### Environment Variables

  

Create or update the `.env` file in the project root:

  

```env

DATABASE_URL=postgres://username:password@localhost:5432/urlshortener?sslmode=disable

```

  

**Configuration Options:**

  

| Variable | Description | Default |

|----------|-------------|----------|

|  `DATABASE_URL`  | PostgreSQL connection string |  `postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable`  |

  

**Connection String Format:**

```

postgres://[username]:[password]@[host]:[port]/[database]?sslmode=[mode]

```

  

### Database Connection Examples

  

**Local Development:**

```

postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable

```

  

**Production (with SSL):**

```

postgres://user:pass@prod-db.example.com:5432/urlshortener?sslmode=require

```

  

## 🏃 Running the Service

  

### Development Mode

  

```bash

# Run directly with Go

go  run  server.go

```

  

### Production Build

  

```bash

# Build the binary

go  build  -o  shortlink-server  server.go

  

# Run the binary

./shortlink-server

```

  

### With Environment Variables

  

```bash

# Linux/macOS

export DATABASE_URL="postgres://user:pass@localhost:5432/urlshortener?sslmode=disable"

go  run  server.go

  

# Windows (PowerShell)

$env:DATABASE_URL="postgres://user:pass@localhost:5432/urlshortener?sslmode=disable"

go  run  server.go

  

# Windows (CMD)

set  DATABASE_URL=postgres://user:pass@localhost:5432/urlshortener?sslmode=disable

go  run  server.go

```

  

The server will start on `http://localhost:8080`

  

## 📚 API Documentation

  

### Base URL

```

http://localhost:8080

```

  

### Endpoints

  

#### 1. Health Check

  

Check if the service is running.

  

**Request:**

```http

GET /health

```

  

**Response:**

```json

{

"status":  "ok"

}

```

  

---

  

#### 2. Create Short URL

  

Shorten a long URL.

  

**Request:**

```http

POST /shorten

Content-Type: application/json

  

{

"url": "https://www.example.com/very/long/url/path"

}

```

  

**Response:**

```json

{

"short_code":  "3dE",

"short_url":  "http://localhost:8080/3dE"

}

```

  

**Status Codes:**

-  `201 Created` - Short URL created successfully

-  `400 Bad Request` - Invalid request body or missing URL

-  `500 Internal Server Error` - Database or server error

  

---

  

#### 3. Redirect to Original URL

  

Access a short URL and get redirected to the original URL.

  

**Request:**

```http

GET /:shortCode

```

  

**Example:**

```http

GET /3dE

```

  

**Response:**

-  `301 Moved Permanently` - Redirects to the original URL

-  `404 Not Found` - Short code doesn't exist

  

```json

{

"message":  "Short URL not found"

}

```

  

---

  

#### 4. Get URL Statistics

  

Retrieve information about a shortened URL.

  

**Request:**

```http

GET /api/stats/:shortCode

```

  

**Example:**

```http

GET /api/stats/3dE

```

  

**Response:**

```json

{

"id":  15432,

"short_code":  "3dE",

"original_url":  "https://www.example.com/very/long/url/path"

}

```

  

**Status Codes:**

-  `200 OK` - URL information retrieved

-  `404 Not Found` - Short code doesn't exist

-  `500 Internal Server Error` - Database error

  

## 💡 Usage Examples

  

### Using cURL

  

**Create a short URL:**

```bash

curl  -X  POST  http://localhost:8080/shorten \

-H "Content-Type: application/json" \

-d  '{"url":"https://github.com/your-username/your-repo"}'

```

  

**Get URL stats:**

```bash

curl  http://localhost:8080/api/stats/3dE

```

  

**Test redirect (with verbose output):**

```bash

curl  -L  -v  http://localhost:8080/3dE

```

  

### Using JavaScript (Fetch API)

  

```javascript

// Create short URL

async  function  shortenURL(longUrl)  {

const  response  =  await  fetch('http://localhost:8080/shorten',  {

method:  'POST',

headers:  {

'Content-Type':  'application/json',

},

body:  JSON.stringify({ url:  longUrl  }),

});

const  data  =  await  response.json();

console.log('Short URL:',  data.short_url);

return  data;

}

  

// Get URL stats

async  function  getStats(shortCode)  {

const  response  =  await  fetch(`http://localhost:8080/api/stats/${shortCode}`);

const  data  =  await  response.json();

console.log('URL Info:',  data);

return  data;

}

  

// Usage

shortenURL('https://example.com/long/url');

getStats('3dE');

```

  

### Using Python (requests)

  

```python

import requests

  

# Create short URL

def  shorten_url(long_url):

response = requests.post(

'http://localhost:8080/shorten',

json={'url': long_url}

)

data = response.json()

print(f"Short URL: {data['short_url']}")

return data

  

# Get URL stats

def  get_stats(short_code):

response = requests.get(f'http://localhost:8080/api/stats/{short_code}')

data = response.json()

print(f"URL Info: {data}")

return data

  

# Usage

shorten_url('https://example.com/long/url')

get_stats('3dE')

```

  

## 🗄️ Database Schema

  

The service automatically creates the following schema on startup:

  

```sql

CREATE  TABLE  IF  NOT  EXISTS urls (

id SERIAL  PRIMARY KEY, -- Auto-incrementing ID

short_code VARCHAR(20) UNIQUE  NOT NULL, -- The Base62 code

original_url TEXT  NOT NULL, -- The original long URL

created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP -- Creation timestamp

);

  

-- Index for faster lookups

CREATE  INDEX  IF  NOT  EXISTS idx_short_code ON urls(short_code);

```

  

**Table: `urls`**

  

| Column | Type | Description |

|--------|------|-------------|

|  `id`  | SERIAL | Auto-incrementing primary key |

|  `short_code`  | VARCHAR(20) | Unique Base62 encoded short code |

|  `original_url`  | TEXT | The original long URL |

|  `created_at`  | TIMESTAMP | When the URL was created |

  

## 🔍 How It Works

  

### Base62 Encoding Algorithm

  

The service uses Base62 encoding to convert sequential database IDs into short codes:

  

1.  **Get Next ID**: Retrieve the next sequential ID from PostgreSQL's `SERIAL` sequence

2.  **Encode to Base62**: Convert the numeric ID to a Base62 string using characters `0-9a-zA-Z`

3.  **Store Mapping**: Save the short code and original URL to the database

4.  **Return Short URL**: Provide the complete shortened URL to the user

  

**Base62 Character Set:**

```

0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ

```

  

**Encoding Examples:**

- ID `0` → `"0"`

- ID `61` → `"Z"`

- ID `62` → `"10"`

- ID `15432` → `"3dE"`

- ID `1000000` → `"4c92"`

  

**Benefits:**

-  **Compact**: Much shorter than numeric IDs

-  **URL-Safe**: Only alphanumeric characters

-  **Sequential**: Predictable growth pattern

-  **Reversible**: Can decode back to ID if needed

  

### Request Flow

  

**Creating a Short URL:**

```

1. Client sends POST /shorten with original URL

2. Server gets next ID from database sequence

3. Server encodes ID to Base62 short code

4. Server saves mapping to database

5. Server returns short code and full short URL

```

  

**Accessing a Short URL:**

```

1. Client visits GET /:shortCode

2. Server looks up short code in database

3. Server retrieves original URL

4. Server sends 301 redirect to original URL

5. Browser follows redirect to destination

```

  

## 🔧 Development

  

### Project Structure

  

```

shortlink-url/

├── server.go # Main application code

├── go.mod # Go module dependencies

├── go.sum # Dependency checksums

├── .env # Environment configuration

└── README.md # This file

```

  

### Adding Features

  

Common enhancements you might want to add:

  

-  **Click Tracking**: Add a `clicks` counter to track URL usage

-  **Expiration**: Add TTL for temporary short URLs

-  **Custom Aliases**: Allow users to specify custom short codes

-  **Analytics**: Track click timestamps, referrers, user agents

-  **Rate Limiting**: Prevent abuse with request throttling

-  **Authentication**: Add API keys or JWT authentication

-  **QR Codes**: Generate QR codes for short URLs

  
  

## 🐛 Troubleshooting

  

### Common Issues

  

**Database Connection Failed:**

```

Failed to connect to database: connection refused

```

**Solution:** Ensure PostgreSQL is running and connection string is correct.

  

```bash

# Check if PostgreSQL is running

sudo  systemctl  status  postgresql

  

# Start PostgreSQL

sudo  systemctl  start  postgresql

```

  

**Port Already in Use:**

```

bind: address already in use

```

**Solution:** Change the port in `server.go` or kill the process using port 8080.

  

```bash

# Find process using port 8080

lsof  -i  :8080

  

# Kill the process

kill  -9  <PID>

```

  

**Module Not Found:**

```

package github.com/labstack/echo/v4: cannot find package

```

**Solution:** Run `go mod download` to install dependencies.

  

**Database Schema Not Created:**

```

relation "urls" does not exist

```

**Solution:** The schema is auto-created on startup. Check database permissions.

  

### Logging

  

The service logs important events:

  

- ✅ Database connected successfully

- ✅ Database schema initialized

- 🚀 Server starting on http://localhost:8080

- HTTP request logs (via Echo middleware)

- Error logs for failed operations

  

### Debug Mode

  

Enable verbose logging by modifying the Echo logger:

  

```go

e.Logger.SetLevel(log.DEBUG)

```

  

## 📄 License

  

This project is open source and available under the [MIT License](LICENSE).

  

## 🤝 Contributing

  

Contributions are welcome! Please feel free to submit a Pull Request.

  

1. Fork the repository

2. Create your feature branch (`git checkout -b feature/amazing-feature`)

3. Commit your changes (`git commit -m 'Add some amazing feature'`)

4. Push to the branch (`git push origin feature/amazing-feature`)

5. Open a Pull Request

  

## 📧 Support

  

If you have any questions or run into issues, please open an issue on GitHub.

  

---

  

**Built with ❤️ using Go and PostgreSQL**