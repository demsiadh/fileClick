#!/bin/bash

# 停止并删除后端容器
if docker stop fileclick-backend &> /dev/null; then
  docker rm fileclick-backend &> /dev/null
fi

# 停止并删除前端容器
if docker stop fileclick-frontend &> /dev/null; then
  docker rm fileclick-frontend &> /dev/null
fi

# 删除后端和前端镜像
docker rmi fileclick-backend fileclick-frontend &> /dev/null || true

# 清理无用镜像
docker image prune -f