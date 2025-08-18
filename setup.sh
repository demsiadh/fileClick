#!/bin/bash

# 构建并运行后端容器
docker build -t fileclick-backend .
docker run -d --name fileclick-backend -p 8080:8080 -v /data/home/bellkang/ownspace/fileClick/data:/data fileclick-backend

# 构建并运行前端容器
cd static && docker build -t fileclick-frontend .
docker run -d --name fileclick-frontend -p 80:80 fileclick-frontend