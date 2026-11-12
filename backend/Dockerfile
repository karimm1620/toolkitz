FROM golang:1.26.5-alpine AS builder

# go-webp memakai cgo -> butuh gcc, musl-dev, dan header libwebp
RUN apk add --no-cache gcc musl-dev libwebp-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o pdf-tool ./cmd/cli

FROM alpine:latest

# libwebp      : runtime library untuk encoder WEBP (cgo, dynamic link)
# libreoffice-*: dipakai endpoint /api/convert/* lewat perintah `soffice`.
#                TANPA ini Word<->PDF selalu gagal (executable not found).
# font-*       : tanpa font, teks hasil konversi jadi kotak/hilang.
RUN apk add --no-cache \
    libwebp \
    libreoffice-common \
    libreoffice-writer \
    libreoffice-draw \
    font-liberation \
    ttf-dejavu

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/pdf-tool .

EXPOSE 3000

CMD ["./server"]
