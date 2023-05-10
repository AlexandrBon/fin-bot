FROM alpine
ENV LANGUAGE="en"
COPY . ../tgbot/internal
RUN apk add --no-cache ca-certificates
EXPOSE 80/tcp
CMD [ "go ./tgbot/internal/cmd/main.go" ]