FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o doctor_service

# Expose the application port
EXPOSE 8083

# Start the service
CMD ["./doctor_service"]
