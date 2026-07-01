#################################
# PersonalMediaVault Dockerfile #
#################################

# Build backend

FROM golang:latest AS backend_builder

    ## Copy backend SSE library
    RUN mkdir -p /root/semantic-search-engine/protocol
    ADD semantic-search-engine/protocol/sse-proto-go /root/semantic-search-engine/protocol/sse-proto-go

    ## Copy backend
    ADD backend /root/backend

    ## Compile backend
    WORKDIR /root/backend
    RUN go build -o pmvd

# Build frontend

FROM node:alpine AS frontend_builder

    ## Copy frontend
    ADD frontend /root/frontend

    ## Build frontend
    WORKDIR /root/frontend
    RUN npm install
    RUN npm run build

# Build semantic search engine server

FROM rust:latest AS sse_server_builder

    ## Install dependencies
    RUN apt update
    RUN apt install -y protobuf-compiler

    ## Copy source code
    ADD semantic-search-engine/server /root/server

    ## Copy protocol files
    RUN mkdir -p protocol
    ADD semantic-search-engine/protocol/sse.proto /root/protocol/sse.proto

    ## Compile
    WORKDIR /root/server
    RUN cargo build --release

# Prepare runner

FROM debian AS runner

    ## Update dependencies sources
    RUN apt update

    ## Install FFMPEG
    RUN apt install -y ffmpeg
    ENV FFMPEG_PATH=/usr/bin/ffmpeg
    ENV FFPROBE_PATH=/usr/bin/ffprobe

    ## Copy backend binary
    COPY --from=backend_builder /root/backend/pmvd /usr/bin/pmvd

    ## Copy SSE server binary
    COPY --from=sse_server_builder /root/server/target/release/pmv-sse /usr/bin/pmv-sse
    ENV SSE_BIN_PATH=/usr/bin/pmv-sse

    ## Copy frontend
    RUN mkdir -p /usr/lib/pmv/
    COPY --from=frontend_builder /root/frontend/dist /usr/lib/pmv/frontend
    ENV FRONTEND_PATH=/usr/lib/pmv/frontend

    ## Working directory
    WORKDIR /root

    ## Default vault folder
    RUN mkdir /vault

    ## Ports
    EXPOSE 80
    EXPOSE 443

    ## Entry point
    ENTRYPOINT ["/usr/bin/pmvd"]
    CMD ["--daemon", "--vault-path", "/vault"]
