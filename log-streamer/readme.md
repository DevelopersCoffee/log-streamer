# Log Streaming Application

This project is a log streaming application built with React for the frontend and Go for the backend. It allows users to stream multiple log files dynamically, view them in a single window, and close individual streams.

## Features

- Dynamically stream multiple log files
- Close individual log streams
- Adaptive layout that adjusts based on screen size
- Supports both JSON and plain text log formats
- Auto-scroll functionality
- Download logs

## Prerequisites

- Node.js (v16 or higher)
- Go (v1.16 or higher)
- A directory `/tmp/local` containing the log files you want to stream

## Setup

### Frontend

1. Navigate to the `frontend` directory:

```sh
cd frontend
```

Install the dependencies:
```
npm install
```

Build the frontend:
npm run build
Backend
Navigate to the project root directory where the main.go file is located.

Build the Go binary:

```sh
go build -o log-streamer -buildvcs=false
```
Running the Application
Start the Go server:

```sh
./log-streamer
```
Open your browser and navigate to:
```arduino
http://localhost:8080
```
API Endpoints
List Log Files

bash
```
GET /api/files
```
Returns a JSON array of log files available in the /tmp/local directory.

Stream Logs
```bash
GET /api/logs/:filename
```
Streams the specified log file. Replace :filename with the actual log file name.

Download Log File
```bash
GET /api/download/:filename
```
Downloads the specified log file. Replace :filename with the actual log file name.

Project Structure
```go

.
├── frontend
│   ├── public
│   │   ├── favicon.ico
│   │   ├── index.html
│   │   ├── logo192.png
│   │   ├── logo512.png
│   │   └── manifest.json
│   ├── src
│   │   ├── App.css
│   │   ├── App.js
│   │   ├── index.css
│   │   └── index.js
│   ├── package.json
│   └── README.md
├── static
│   ├── static
│   │   ├── css
│   │   ├── js
│   │   └── ...
│   ├── favicon.ico
│   ├── index.html
│   └── manifest.json
├── main.go
├── go.mod
├── go.sum
└── README.md
```
Usage:

```sh
make all
```
TODO
- Search logs
- Spring Boot formatting
- Key-based filters
- kafka events to a subscriber and then to UI via SSE is powerful for one way communication without use of websocket.
