FROM scratch
#FROM golang:1.6
MAINTAINER Tom Scanlan <tscanlan@vmware.com>

EXPOSE 9998

# Add the microservice
ADD q3reservation /q3reservation

CMD ["/q3reservation", "--port", "9998"]
