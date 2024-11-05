# pulse/Dockerfile
FROM golang:1.17-alpine

WORKDIR /app

# Copy project files and build the app
COPY . .
RUN go mod download
RUN go build -o pulse-server

# Create an empty log file
RUN touch pulse_logs.txt

# Start the application
CMD ["./pulse-server"]
