#################################
# PersonalMediaVault Dockerfile #
#################################

# Build backend

FROM golang:alpine AS backend_builder

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

# Prepare runner

FROM alpine AS runner

    ## Install common libraries
    RUN apk add gcompat

    ## Install FFMPEG
    RUN apk add --no-cache ffmpeg
    ENV FFMPEG_PATH=/usr/bin/ffmpeg
    ENV FFPROBE_PATH=/usr/bin/ffprobe

    ## Copy backend binary
    COPY --from=backend_builder /root/backend/pmvd /usr/bin/pmvd

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
