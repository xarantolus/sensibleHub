FROM golang:1.18 as builder

# Build the normal executable
RUN mkdir /build
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -a -v -mod vendor -ldflags "-s -w" -o sensibleHub .


# Now for the image we actually run the server in
FROM alpine:latest
RUN apk add ca-certificates ffmpeg python3
# Copy main executable
COPY --from=builder /build/sensibleHub .
# Download yt-dlp
RUN wget -qO /bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && chmod +x /bin/yt-dlp
ENV PATH="/bin:${PATH}"
ENV RUNNING_IN_DOCKER=true
ENTRYPOINT [ "./sensibleHub", "-config", "/config/config.json" ]
