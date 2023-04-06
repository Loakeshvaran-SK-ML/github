# Using official golang docker image 
FROM golang:1.20.2-alpine

# Setting working directory
WORKDIR /app

# Creating a Arguments for USERNAME and PASSWORD
ARG USERNAME
ARG PASSWORD 

# Creating Environment variables for USERNAME and PASSWORD
ENV USERNAME=$USERNAME
ENV PASSWORD=$PASSWORD

# Copying source code to work directory
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY parser ./

# Creating a config volume
VOLUME /config
COPY ./config/config.yaml /config/config.yaml

# Downloading required packages
RUN go mod download

# Building the binary file
RUN go build -o /github-parser

# Running the command to start the init Container
ENTRYPOINT ["/github-parser"]
