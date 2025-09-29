# GoHTML Server

A self-hosted **HTML → PDF conversion server**, fully compatible with the [gohtml client library].  
This project serves as a **drop-in replacement** for UniDoc’s commercial UniHTML server.

---

## ✨ Features
- Convert **HTML files** to PDF  
- Convert **HTML directories** (with CSS, JS, images) to PDF  
- Convert **web URLs** to PDF  
- Customizable **page size** (A4, Letter, custom dimensions)  
- Configurable **margins** and **orientation**  
- Support for **dynamic content rendering** (JavaScript)  
- Wait for **specific selectors** (ready/visible)  
- **Pixel-perfect output** via Chrome/Chromium rendering engine  
- **Docker support** for easy deployment  

---

## 🏗 Architecture
This project includes two main components:

1. **Server** (`server/main.go`) – HTTP server that converts HTML to PDF using [chromedp].  
2. **Client** – the [gohtml] library that communicates with the server.  

---

## 📦 Prerequisites

### Option 1: Docker (recommended)
- Docker installed on your system  

### Option 2: Local development
- Go **1.21+**  
- Chrome/Chromium (will be auto-downloaded by `chromedp` if missing)  

---

## 🚀 Quick Start with Docker

### Build image
```bash
docker build -t gohtml-server:latest .
```

### Run server

```bash
docker run -p 8080:8080 gohtml-server:latest
```

### Run in background

```bash
docker run -d -p 8080:8080 --name gohtml-server gohtml-server:latest
```

```bash
[gohtml]: https://github.com/unitechio/gohtml
[chromedp]: https://github.com/chromedp/chromedp
```

### Docker Compose

```bash
# Start server
docker-compose up -d

# View logs
docker-compose logs -f

# Stop server
docker-compose down
```

## Local Development

### Install Dependencies

```bash
cd server
go mod download
```

### Run Server

```bash
go run ./server/main.go
```

Server will start on `http://localhost:8080`

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

### Convert HTML to PDF

```bash
POST http://localhost:8080/v1/pdf
Content-Type: application/json

{
  "content": "base64_encoded_html_or_raw_bytes",
  "contentType": "text/html",
  "method": "html",
  "PageParameters": {
    "paperWidth": "210mm",
    "paperHeight": "297mm",
    "orientation": "portrait",
    "marginTop": "10mm",
    "marginBottom": "10mm",
    "marginLeft": "10mm",
    "marginRight": "10mm"
  },
  "RenderParameters": {
    "waitTime": 1000
  }
}
```

**Response**: PDF binary data with status `201 Created`

## Using with gohtml Client

### Install Client Library

```bash
go get github.com/unitechio/gohtml
go get github.com/unitechio/gopdf
```

### Example Code

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/unitechio/gohtml"
    "github.com/unitechio/gopdf/creator"
    "github.com/unitechio/gopdf/common"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Println("Usage: program <server-url> <input-path>")
        os.Exit(1)
    }

    // Connect to server
    if err := gohtml.Connect(os.Args[1]); err != nil {
        fmt.Printf("Connect failed: %v\n", err)
        os.Exit(1)
    }

    // Create PDF creator
    c := creator.New()

    // Load HTML document
    doc, err := gohtml.NewDocument(os.Args[2])
    if err != nil {
        fmt.Printf("NewDocument failed: %v\n", err)
        os.Exit(1)
    }

    // Convert to PDF
    if err = c.Draw(doc); err != nil {
        fmt.Printf("Draw failed: %v\n", err)
        os.Exit(1)
    }

    // Write output
    if err = c.WriteToFile("output.pdf"); err != nil {
        fmt.Printf("WriteToFile failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Success: output.pdf created")
}
```

### Run Client

```bash
# Convert HTML file
go run main.go http://localhost:8080 input.html

# Convert web page
go run main.go http://localhost:8080 https://example.com

# Convert directory
go run main.go http://localhost:8080 ./html-folder
```

## Configuration

### Page Sizes

Supported page sizes:
- `A4` (210mm × 297mm)
- `Letter` (8.5in × 11in)
- Custom dimensions (e.g., `paperWidth: "200mm", paperHeight: "300mm"`)

### Margins

Default margins: 10mm on all sides

```go
doc.SetMargins(left, right, top, bottom) // in points
```

### Orientation

```go
doc.SetLandscapeOrientation() // Default is portrait
```

### Wait for Content

```go
// Wait for time
doc.WaitTime(2 * time.Second)

// Wait for selector to be ready
doc.WaitReady(".content", selector.ByCSS)

// Wait for selector to be visible
doc.WaitVisible("#main", selector.ByID)
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `TZ` | Timezone | UTC |

## Performance

- Typical conversion time: 2-5 seconds per page
- Memory usage: ~200MB per Chrome instance
- Concurrent requests: Limited by available RAM

## Troubleshooting

### Chrome fails to start

**Problem**: `chromedp run: chrome failed to start`

**Solution**: 
- Use Docker (recommended) - Chrome is pre-installed
- Or let chromedp auto-download Chromium (~150MB)

### Out of memory

**Problem**: Server crashes under load

**Solution**:
```bash
# Limit concurrent conversions
# Add connection pooling
# Increase Docker memory limit
docker run --memory=2g -p 8080:8080 gohtml-server
```

### Timeout errors

**Problem**: Conversion takes too long

**Solution**:
```go
doc.SetTimeoutDuration(60 * time.Second) // Increase timeout
```

## Project Structure

```
.
├── server/
│   └── main.go           # Server implementation
├── Dockerfile            # Docker build configuration
├── docker-compose.yml    # Docker Compose setup
├── go.mod               # Go dependencies
├── go.sum
└── README.md
```

## Technology Stack

- **Language**: Go 1.21+
- **PDF Engine**: chromedp (Chrome DevTools Protocol)
- **HTTP Server**: net/http (stdlib)
- **Container**: Docker + chromedp/headless-shell

## License

This is a fork of UniDoc's UniHTML with license restrictions removed. Use at your own discretion.

## Credits

- Original UniHTML: [UniDoc.io](https://unidoc.io)
- chromedp: [github.com/chromedp/chromedp](https://github.com/chromedp/chromedp)
- gohtml fork: [github.com/unitechio/gohtml](https://github.com/unitechio/gohtml)

## Support

For issues and questions, please open an issue on GitHub.

## Alternatives

- **wkhtmltopdf**: Faster but outdated WebKit engine
- **Puppeteer**: Node.js alternative with similar features
- **WeasyPrint**: Python-based, CSS-focused
- **UniHTML Cloud**: Official commercial service
```