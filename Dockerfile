FROM golang:1.14.3 AS builder
WORKDIR /app

COPY go.mod go.sum *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cardsagainsthumanity github.com/bcspragu/OnlineCAH

FROM alpine:3.7
WORKDIR /app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=builder /app/cardsagainsthumanity /app/
COPY css/ css
COPY js/ js/
COPY img/ img/
COPY templates/ templates/
COPY cards.json cards.json
COPY players.json players.json
COPY code code

CMD ["./cardsagainsthumanity"]
