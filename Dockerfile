FROM node:24-alpine AS frontend

RUN apk add --no-cache make
WORKDIR /src
COPY Makefile ./
COPY frontend/package.json frontend/package-lock.json frontend/
RUN make frontend-deps
COPY frontend/ frontend/
RUN make frontend-build


FROM golang:1-alpine AS builder

RUN apk add --no-cache make
WORKDIR /build
COPY . /build
COPY --from=frontend /src/frontend/dist /build/frontend/dist
RUN CGO_ENABLED=0 make server


# Now for the image we actually run the server in
FROM alpine:latest
RUN apk add ca-certificates ffmpeg python3 curl
# Copy main executable
COPY --from=builder /build/sensibleHub .
# Download yt-dlp

RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
	chmod a+rx /usr/local/bin/yt-dlp  && \
	ln /usr/local/bin/yt-dlp /usr/local/bin/youtube-dl && \
	ln /usr/local/bin/yt-dlp /usr/local/bin/youtube-dlp && \
	ln /usr/local/bin/yt-dlp /usr/local/bin/youtube-dlc && \
	ln /usr/local/bin/yt-dlp /usr/local/bin/yt-dlc

ENV PATH="/bin:${PATH}"
ENV RUNNING_IN_DOCKER=true
ENTRYPOINT [ "./sensibleHub", "-config", "/config/config.json" ]
