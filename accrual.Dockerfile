FROM golang:1.22
COPY ./cmd/accrual/. /usr/local/bin/
CMD ["accrual_linux_amd64"]
