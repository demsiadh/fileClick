# 第一阶段：构建前端
FROM alpine:latest

# 安装Nginx
RUN apk add --no-cache nginx

# 复制前端静态文件
COPY index.html /usr/share/nginx/html/index.html

# 配置Nginx
COPY nginx.conf /etc/nginx/nginx.conf

# 暴露前端端口
EXPOSE 80

# 启动Nginx
CMD nginx -g 'daemon off;'