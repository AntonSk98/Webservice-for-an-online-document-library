FROM alpine:latest

RUN apk add --update nginx && rm  -rf /var/cache/apk/*
RUN apk add nginx

COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx"]
