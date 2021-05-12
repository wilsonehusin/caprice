FROM gcr.io/distroless/static:nonroot

COPY _output/caprice /usr/local/bin/

ENTRYPOINT ["caprice"]
