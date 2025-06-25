FROM golang:alpine


RUN apk add --no-cache bash build-base
WORKDIR /app

CMD ["bash"]