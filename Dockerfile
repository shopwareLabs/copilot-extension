FROM chainguard/go AS builder

WORKDIR /app
COPY . .

RUN go build -ldflags="-s -w" -trimpath -o /rag

FROM chainguard/wolfi-base

RUN apk add --no-cache ca-certificates

COPY --from=builder /rag /app/rag
COPY --from=builder /app/db /app/db

EXPOSE 8000

WORKDIR /app
ENTRYPOINT ["/app/rag"]
CMD [ "server" ]
