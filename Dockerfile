FROM golang:1.18 as builder

# Build the normal executable
RUN mkdir /build
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -mod vendor -ldflags "-s -w" -o sensibleHub .


# Now for the image we actually run the server in
FROM python:3-alpine

# Install ffmpeg
RUN apk add ca-certificates ffmpeg
ENV PATH="/bin:${PATH}"

RUN apk add --no-cache --virtual .pynacl_deps build-base python3-dev libffi-dev

# Install youtube-dl
RUN pip install yt-dlp

RUN apk del .pynacl_deps

# Copy main executable
COPY --from=builder /build/sensibleHub .
ENTRYPOINT [ "./sensibleHub", "-config", "/config/config.json" ]
