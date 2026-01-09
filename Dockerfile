# 多阶段构建 - 前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# 多阶段构建 - 后端
FROM golang:1.21-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o confighub ./cmd/server

# 最终镜像
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制后端二进制
COPY --from=backend-builder /app/confighub .

# 复制前端静态文件
COPY --from=frontend-builder /app/web/dist ./static

# 复制配置文件模板
COPY config.yaml.example ./config.yaml

# 复制数据库迁移文件
COPY migrations ./migrations

EXPOSE 8080

ENV TZ=Asia/Shanghai

CMD ["./confighub"]
