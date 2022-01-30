#------------------------------------------------------------------------------
# Building the base image with dependencies installed
#------------------------------------------------------------------------------
FROM golang as build

# Install dependencies
RUN apt-get update
RUN apt-get install -y iputils-ping iproute2

# Run all commands from the application directory
WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Add all of our code
COPY . /app
