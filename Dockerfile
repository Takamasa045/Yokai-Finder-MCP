FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
