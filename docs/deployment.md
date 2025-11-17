# Go-Uni-Pay 部署指南

## 部署前准备

### 1. 环境要求

- **服务器**: Linux (推荐 Ubuntu 20.04+)
- **Go**: 1.21+
- **MySQL**: 8.0+
- **Redis**: 6.0+
- **内存**: 最少 2GB
- **磁盘**: 最少 10GB

### 2. 安装依赖软件

#### 安装 MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# 启动 MySQL
sudo systemctl start mysql
sudo systemctl enable mysql

# 设置 root 密码
sudo mysql_secure_installation
```

#### 安装 Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server

# 启动 Redis
sudo systemctl start redis
sudo systemctl enable redis
```

## 部署方式

### 方式一：源码部署

#### 1. 克隆代码

```bash
git clone https://github.com/zqdfound/go-uni-pay.git
cd go-uni-pay
```

#### 2. 安装 Go 依赖

```bash
go mod tidy
```

#### 3. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE uni_pay DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p uni_pay < scripts/init.sql
```

#### 4. 修改配置文件

```bash
cp configs/config.yaml configs/config.production.yaml
vim configs/config.production.yaml
```

修改以下配置项：

```yaml
server:
  mode: release  # 生产环境

database:
  host: localhost
  port: 3306
  username: root
  password: your_mysql_password
  database: uni_pay

redis:
  host: localhost
  port: 6379
  password: your_redis_password  # 如果有密码

jwt:
  secret: your-production-secret-key  # 生产环境必须修改
```

#### 5. 编译项目

```bash
make build
```

#### 6. 启动服务

```bash
# 方式1: 直接运行
./bin/go-uni-pay configs/config.production.yaml

# 方式2: 使用 systemd 管理
sudo cp scripts/go-uni-pay.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start go-uni-pay
sudo systemctl enable go-uni-pay
```

### 方式二：Docker 部署

#### 1. 使用 Docker Compose（推荐）

```bash
# 克隆代码
git clone https://github.com/zqdfound/go-uni-pay.git
cd go-uni-pay

# 修改配置
vim configs/config.yaml

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app
```

#### 2. 单独构建 Docker 镜像

```bash
# 构建镜像
docker build -t go-uni-pay:latest .

# 运行容器
docker run -d \
  --name go-uni-pay \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/logs:/app/logs \
  go-uni-pay:latest
```

## 生产环境配置

### 1. Nginx 反向代理

创建 Nginx 配置文件 `/etc/nginx/sites-available/go-uni-pay`:

```nginx
upstream go_uni_pay {
    server 127.0.0.1:8080;
    # 多节点部署
    # server 127.0.0.1:8081;
    # server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name pay.your-domain.com;

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name pay.your-domain.com;

    # SSL 证书
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 日志
    access_log /var/log/nginx/go-uni-pay-access.log;
    error_log /var/log/nginx/go-uni-pay-error.log;

    location / {
        proxy_pass http://go_uni_pay;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 限流
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://go_uni_pay;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/go-uni-pay /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 2. Systemd 服务配置

创建文件 `/etc/systemd/system/go-uni-pay.service`:

```ini
[Unit]
Description=Go Uni Pay Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/go-uni-pay
ExecStart=/opt/go-uni-pay/bin/go-uni-pay /opt/go-uni-pay/configs/config.production.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# 环境变量
Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
```

管理服务：

```bash
# 重新加载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start go-uni-pay

# 开机自启
sudo systemctl enable go-uni-pay

# 查看状态
sudo systemctl status go-uni-pay

# 查看日志
sudo journalctl -u go-uni-pay -f
```

### 3. 日志轮转

创建文件 `/etc/logrotate.d/go-uni-pay`:

```
/opt/go-uni-pay/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 www-data www-data
    sharedscripts
    postrotate
        systemctl reload go-uni-pay > /dev/null 2>&1 || true
    endscript
}
```

## 多节点部署

### 1. 部署多个实例

在不同端口启动多个实例：

```bash
# 节点1 (端口 8080)
./bin/go-uni-pay configs/config.production.yaml

# 节点2 (端口 8081)
cp configs/config.production.yaml configs/config.node2.yaml
# 修改 port: 8081
./bin/go-uni-pay configs/config.node2.yaml

# 节点3 (端口 8082)
cp configs/config.production.yaml configs/config.node3.yaml
# 修改 port: 8082
./bin/go-uni-pay configs/config.node3.yaml
```

### 2. Nginx 负载均衡

```nginx
upstream go_uni_pay {
    # 轮询
    server 127.0.0.1:8080 weight=1;
    server 127.0.0.1:8081 weight=1;
    server 127.0.0.1:8082 weight=1;

    # 健康检查
    keepalive 32;
}
```

## 监控和维护

### 1. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health

# 响应应该是：{"status":"ok"}
```

### 2. 性能监控

推荐使用以下工具：

- **Prometheus + Grafana**: 监控服务指标
- **ELK Stack**: 日志分析
- **Sentry**: 错误追踪

### 3. 数据库备份

```bash
# 创建备份脚本
cat > /opt/backup/backup-mysql.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backup/mysql"
mkdir -p $BACKUP_DIR

mysqldump -u root -p'your_password' uni_pay > $BACKUP_DIR/uni_pay_$DATE.sql
gzip $BACKUP_DIR/uni_pay_$DATE.sql

# 删除30天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete
EOF

chmod +x /opt/backup/backup-mysql.sh

# 设置定时任务
crontab -e
# 添加：每天凌晨2点备份
0 2 * * * /opt/backup/backup-mysql.sh
```

## 安全加固

### 1. 防火墙设置

```bash
# 只允许特定端口
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

### 2. MySQL 安全

```sql
-- 创建专用用户
CREATE USER 'unipay'@'localhost' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON uni_pay.* TO 'unipay'@'localhost';
FLUSH PRIVILEGES;
```

### 3. Redis 安全

编辑 `/etc/redis/redis.conf`:

```
# 设置密码
requirepass your_strong_password

# 绑定本地
bind 127.0.0.1

# 禁用危险命令
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

重启 Redis:

```bash
sudo systemctl restart redis
```

## 故障排查

### 1. 查看日志

```bash
# 应用日志
tail -f logs/app.log

# 系统日志
sudo journalctl -u go-uni-pay -f

# Nginx 日志
tail -f /var/log/nginx/go-uni-pay-error.log
```

### 2. 常见问题

#### 服务无法启动

```bash
# 检查配置文件
./bin/go-uni-pay --config configs/config.production.yaml

# 检查端口占用
sudo lsof -i :8080

# 检查数据库连接
mysql -h localhost -u root -p uni_pay
```

#### 数据库连接失败

- 检查 MySQL 是否运行：`sudo systemctl status mysql`
- 检查用户名密码是否正确
- 检查数据库是否存在

#### Redis 连接失败

- 检查 Redis 是否运行：`sudo systemctl status redis`
- 检查密码是否正确
- 检查网络连接

## 性能优化

### 1. MySQL 优化

编辑 `/etc/mysql/mysql.conf.d/mysqld.cnf`:

```ini
[mysqld]
# 连接数
max_connections = 500

# 缓存
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M

# 查询缓存
query_cache_size = 64M
query_cache_limit = 2M
```

### 2. Redis 优化

```ini
# 最大内存
maxmemory 1gb

# 淘汰策略
maxmemory-policy allkeys-lru
```

### 3. 应用优化

- 启用 HTTP/2
- 开启 gzip 压缩
- 使用 CDN 加速
- 配置合理的连接池大小

## 更新部署

```bash
# 拉取最新代码
git pull

# 编译
make build

# 重启服务
sudo systemctl restart go-uni-pay

# 检查状态
sudo systemctl status go-uni-pay
```

## 回滚

```bash
# 切换到上一个版本
git checkout <previous-commit>

# 重新编译
make build

# 重启服务
sudo systemctl restart go-uni-pay
```
