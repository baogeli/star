# Docker + Kubernetes 部署完整流程指南

## 📋 前置准备

### 1. 确保已安装的工具
- Go 1.24+
- Docker
- kubectl
- kustomize

### 2. 登录阿里云镜像仓库
```bash
docker login registry.cn-hangzhou.aliyuncs.com
```

---

## 🚀 第一步：构建应用

### 1.1 进入项目目录
```bash
cd /home/boa/golang/star
```

### 1.2 编译二进制文件
```bash
make build
```

**说明**：这会生成 `temp/linux_amd64/main` 可执行文件

---

## 📦 第二步：构建并推送 Docker 镜像

### 2.1 构建镜像（本地测试）
```bash
make image TAG=v1.0.0
```

**生成的镜像名**：`registry.cn-hangzhou.aliyuncs.com/boa/template-single:v1.0.0`

### 2.2 构建并推送到阿里云
```bash
make image.push TAG=v1.0.0
```

### 2.3 验证镜像
```bash
# 查看本地镜像
docker images | grep template-single

# 查看推送的标签
docker images registry.cn-hangzhou.aliyuncs.com/boa/template-single
```

---

## ⚙️ 第三步：配置 MySQL 连接

### 3.1 修改本地配置文件
编辑 `manifest/config/config.yaml`，将数据库地址改为：
```yaml
database:
  default:
    link: "mysql:root:123456@tcp(host.docker.internal:3306)/star?parseTime=true&loc=Local"
```

**关键点**：
- `host.docker.internal` 是 Docker 访问宿主机的特殊域名
- 如果 MySQL 在其他服务器，改成实际 IP 地址

### 3.2 确认 MySQL 正在运行
```bash
netstat -tuln | grep 3306
```

应该看到：
```
tcp        0      0 0.0.0.0:3306            0.0.0.0:*               LISTEN
```

---

## 🐳 第四步：Docker 启动容器

### 4.1 启动容器（挂载配置文件）
```bash
docker run -d \
  --name star-app \
  -p 8000:8000 \
  --add-host=host.docker.internal:host-gateway \
  -v /home/boa/golang/star/manifest/config:/app/manifest/config \
  registry.cn-hangzhou.aliyuncs.com/boa/template-single:v1.0.0
```

**参数说明**：
- `-d`：后台运行
- `--name star-app`：容器名称
- `-p 8000:8000`：端口映射（宿主机:容器）
- `--add-host=host.docker.internal:host-gateway`：让容器能访问宿主机
- `-v`：挂载配置文件目录

### 4.2 查看容器状态
```bash
# 查看运行中的容器
docker ps

# 查看容器日志
docker logs star-app

# 实时查看日志
docker logs -f star-app
```

### 4.3 测试服务
```bash
# 测试接口是否正常
curl http://127.0.0.1:8000/v1/show

# 测试数据库连接
curl http://127.0.0.1:8000/v1/fundVolumeList
```

### 4.4 管理容器
```bash
# 停止容器
docker stop star-app

# 删除容器
docker rm star-app

# 重启容器
docker restart star-app

# 查看所有容器（包括已停止的）
docker ps -a
```

---

## 🔧 第五步：修改配置（无需重新打包）

### 5.1 修改配置文件
直接编辑 `manifest/config/config.yaml`

### 5.2 重启容器使配置生效
```bash
docker restart star-app
```

**或者**停止后重新启动：
```bash
docker stop star-app && docker rm star-app

docker run -d \
  --name star-app \
  -p 8000:8000 \
  --add-host=host.docker.internal:host-gateway \
  -v /home/boa/golang/star/manifest/config:/app/manifest/config \
  registry.cn-hangzhou.aliyuncs.com/boa/template-single:v1.0.0
```

---

## ☸️ 第六步：Kubernetes 部署（可选）

### 6.1 更新 K8s 配置文件
编辑 `manifest/deploy/kustomize/overlays/develop/configmap-star.yaml`，设置正确的数据库地址：
```yaml
database:
  default:
    link: "mysql:root:123456@tcp(你的MySQL地址:3306)/star?parseTime=true&loc=Local"
```

### 6.2 部署到 K8s
```bash
# 部署应用
make deploy ENV=develop TAG=v1.0.0
```

### 6.3 查看 K8s 状态
```bash
# 查看 Pod 状态
kubectl get pods

# 查看 Deployment
kubectl get deployments

# 查看 Service
kubectl get services

# 查看日志
kubectl logs -f deployment/template-single
```

### 6.4 更新配置后重启
```bash
# 应用新配置
kubectl apply -f manifest/deploy/kustomize/overlays/develop/configmap-star.yaml

# 重启 Pod
kubectl rollout restart deployment/template-single
```

---

## 📝 常用命令速查表

| 操作 | 命令 |
|------|------|
| **构建镜像** | `make image TAG=v1.0.0` |
| **推送镜像** | `make image.push TAG=v1.0.0` |
| **启动容器** | `docker run -d --name star-app -p 8000:8000 --add-host=host.docker.internal:host-gateway -v $(pwd)/manifest/config:/app/manifest/config registry.cn-hangzhou.aliyuncs.com/boa/template-single:v1.0.0` |
| **查看日志** | `docker logs -f star-app` |
| **重启容器** | `docker restart star-app` |
| **停止容器** | `docker stop star-app && docker rm star-app` |
| **测试接口** | `curl http://127.0.0.1:8000/v1/fundVolumeList` |
| **K8s 部署** | `make deploy ENV=develop TAG=v1.0.0` |

---

## ⚠️ 常见问题

### 问题 1：端口被占用
```bash
# 查看哪个进程占用了 8000 端口
lsof -i :8000

# 或者查看 Docker 容器
docker ps | grep 8000

# 停止冲突的容器
docker stop <容器名>
```

### 问题 2：数据库连接失败
- 检查 MySQL 是否运行：`netstat -tuln | grep 3306`
- 检查配置文件中的地址是否正确
- 检查防火墙是否允许连接

### 问题 3：配置文件未生效
- 确认挂载路径正确：`-v /完整路径/manifest/config:/app/manifest/config`
- 重启容器使配置生效

### 问题 4：镜像拉取失败
- 确认已登录：`docker login registry.cn-hangzhou.aliyuncs.com`
- 检查镜像标签是否正确

---

## 🎯 练习建议

按照以下顺序独立练习：

1. ✅ 先练习构建和推送镜像
2. ✅ 再练习本地 Docker 启动和测试
3. ✅ 然后练习修改配置并重启
4. ✅ 最后尝试 K8s 部署（如果有 K8s 环境）

祝你练习顺利！🚀
