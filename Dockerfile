FROM golang:1.19-alpine

WORKDIR /app
COPY . .

RUN go mod tidy && go build -o manager .

EXPOSE 8080

CMD ["./manager"]
