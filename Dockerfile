FROM golang:alpine

WORKDIR /root
RUN mkdir /vault

# Install dependencies

RUN apk add --no-cache ffmpeg
ENV FFMPEG_PATH=/usr/bin/ffmpeg
ENV FFPROBE_PATH=/usr/bin/ffprobe

# Copy backend

ADD backend /root/backend

# Copy frontend

ADD frontend/dist /root/frontend

ENV FRONTEND_PATH=/root/frontend

# Compile backend

WORKDIR /root/backend

RUN go build -o pmvd

WORKDIR /root

# Ports

EXPOSE 80
EXPOSE 443

# Entry point

ENTRYPOINT ["/root/backend/pmvd"]
CMD ["--daemon", "--vault-path", "/vault"]
