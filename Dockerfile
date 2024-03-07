# Start from the latest golang base image
FROM cgr.dev/chainguard/go:latest as build

# Add Maintainer Info
LABEL maintainer="Kwesi Henry"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Start a new stage from scratch to create a minimal final image
FROM cgr.dev/chainguard/debian-base:latest

# Add Maintainer Info
LABEL maintainer="Kwesi Henry"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=build /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Use a non-root user to run our application (ChainGuard best practice)
RUN chown -R nobody:nogroup /app && chmod -R 755 /app
USER nobody

# Command to run the executable
CMD ["./main"]
