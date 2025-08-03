# Use the official Go image as the base image
FROM golang:1.21-alpine

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY backend/ ./backend/

# Build the application
RUN cd backend && go build -o main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./backend/main"] 