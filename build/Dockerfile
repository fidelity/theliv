FROM golang:latest AS build-go

#Setup the proxy and No Proxy
ARG https_proxy
ARG no_proxy
ENV https_proxy=${https_proxy}
ENV no_proxy=${no_proxy}

# Move to working directory /go-service
WORKDIR /go-service
# COPY . /go-service/
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN mkdir -p pkg
COPY pkg/. pkg 
RUN mkdir -p internal
COPY internal/. internal

ENV GOPROXY=direct
# Build the application
RUN go mod tidy -compat=1.17
RUN go build -o main .


# 2) BUILD UI
FROM node:16-alpine3.12 AS build-node

#Setup the proxy and No Proxy
ARG https_proxy
ARG no_proxy
ENV https_proxy=${https_proxy}
ENV no_proxy=${no_proxy}

RUN mkdir /app
WORKDIR /app
COPY website/. .


# Configure npm
RUN npm install 
RUN npm run build


# 3) BUILD FINAL IMAGE
FROM nginx:latest

# change uid & gid of nginx user
RUN usermod -u 10014 nginx && groupmod -g 10014 nginx

# modify ca-certificate file permissions
RUN chown -R nginx:nginx /usr/share/ca-certificates && \
        chown nginx:nginx /etc/ca-certificates.conf && \
        chown -R nginx:nginx /etc/ssl/certs

COPY --from=build-go /go-service/main /app/server/
COPY --from=build-node /app/dist/theliv /app/client/theliv

RUN mkdir /app/client/theliv/static
RUN mv -f /app/client/theliv/*.js /app/client/theliv/*.css /app/client/theliv/static/
RUN find /app/client/theliv/ | grep fa- | xargs -I '{}' mv '{}' /app/client/theliv/static/

COPY web/nginx/default.conf /etc/nginx/conf.d/default-temp.conf
COPY web/nginx/nginx.conf /etc/nginx/nginx-temp.conf
COPY web/nginx/datadog.conf /etc/nginx/conf.d/datadog.conf

RUN chown nginx:nginx /etc/nginx/conf.d/default-temp.conf && \
        chown nginx:nginx /etc/nginx/nginx-temp.conf && \
        chown -R nginx:nginx /etc/nginx/conf.d/datadog.conf

COPY web/nginx/startup.sh /
RUN chown nginx:nginx /startup.sh && \
        chmod ug+x /startup.sh