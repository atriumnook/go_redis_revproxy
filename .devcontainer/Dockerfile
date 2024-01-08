FROM amazonlinux:latest

RUN dnf install -y \
    git \
    golang && \
    dnf clean all

WORKDIR /app

RUN go install -v golang.org/x/tools/gopls@latest && \
    go install -v github.com/go-delve/delve/cmd/dlv@latest

# RUN go mod init redis-revproxy && \
#     go mod tidy

# CMD ["go", "run", "main.go"]
