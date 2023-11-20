FROM alpine:latest

RUN mkdir /app

COPY accountService /app

CMD [ "/app/accountService"]