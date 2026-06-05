# New API 生产环境安全升级指南

本指南用于将现有生产环境中的 New API 服务替换为此重构版本，同时保留 PostgreSQL、Redis、挂载数据、用户、余额、令牌、渠道、充值记录、兑换记录和系统配置。

核心规则很简单：只替换应用镜像或应用代码。不要替换生产数据库、Redis 卷、`.env` 或 Docker 卷。

## 1. 生产资产清单

在操作生产环境之前，先记录当前值：

```bash
cd /opt/new-api
docker compose ps
docker compose config > backup.compose.rendered.$(date +%F-%H%M%S).yml
```

清单检查项：

- 当前 `docker-compose.yml` 路径。
- 当前 `.env` 路径。
- New API 服务名、镜像标签、容器名和挂载的数据目录。
- PostgreSQL 服务/容器名、数据库名、用户名、密码来源和卷名。
- Redis 服务/容器名和卷名。
- 反向代理配置，例如 Nginx/Caddy/Cloudflare tunnel。
- 当前生产镜像标签，用于回滚。

常用命令：

```bash
docker volume ls
docker inspect <postgres_container> --format '{{json .Mounts}}'
docker inspect <redis_container> --format '{{json .Mounts}}'
docker inspect <newapi_container> --format '{{json .Mounts}}'
```

## 2. 升级前备份

创建时间戳和备份目录：

```bash
TS=$(date +%F-%H%M%S)
mkdir -p /opt/new-api-backups/$TS
cd /opt/new-api
```

备份 PostgreSQL：

```bash
docker exec <postgres_container> pg_dump \
  -U <postgres_user> \
  -d <postgres_database> \
  -Fc \
  -f /tmp/newapi-$TS.dump

docker cp <postgres_container>:/tmp/newapi-$TS.dump /opt/new-api-backups/$TS/
```

备份 Redis 持久化数据或卷数据。如果 Redis 在挂载目录中使用 appendonly 文件或 `dump.rdb`：

```bash
cp -a /path/to/redis/data /opt/new-api-backups/$TS/redis-data
```

如果 Redis 使用 Docker 卷：

```bash
docker run --rm \
  -v <redis_volume>:/data:ro \
  -v /opt/new-api-backups/$TS:/backup \
  alpine tar czf /backup/redis-volume.tgz -C /data .
```

备份 compose、env 和应用挂载数据：

```bash
cp -a docker-compose.yml /opt/new-api-backups/$TS/
cp -a .env /opt/new-api-backups/$TS/
cp -a /path/to/new-api/data /opt/new-api-backups/$TS/new-api-data
```

备份和升级期间禁止：

- 不要运行 `docker compose down -v`。
- 不要删除 Docker 卷。
- 不要覆盖生产环境 `.env`。
- 不要初始化一个空的生产数据库。

## 3. 使用生产备份进行预发布验证

先将生产环境 PostgreSQL 备份恢复到测试数据库或测试服务器：

```bash
createdb newapi_upgrade_test
pg_restore -U <postgres_user> -d newapi_upgrade_test /opt/new-api-backups/$TS/newapi-$TS.dump
```

启动新应用，并连接到测试数据库和测试 Redis 实例。使用复制过来的 `.env` 值，但只修改：

- PostgreSQL 主机/数据库，改为测试数据库。
- Redis 主机/数据库，改为测试 Redis。
- 如果测试支付回调，则修改公开回调/域名相关值。

然后启动新镜像：

```bash
docker compose -f docker-compose.upgrade-test.yml up -d
docker compose -f docker-compose.upgrade-test.yml logs -f new-api
```

验证清单：

- 现有用户可以登录。
- 现有余额没有变化。
- 现有 API 令牌仍然可用。
- 现有渠道、模型倍率、分组和配置仍然保留。
- 现有充值和兑换历史仍然保留。
- 现有 Redis 会话/缓存不会导致启动错误。
- 新的分销相关表会自动创建。
- 分销系统默认禁用，直到管理员启用。
- 当上游返回官方 `usage` 时，DeepSeek 调用会记录官方 `usage`。

如果预发布环境中有任何必要检查项失败，不要操作生产环境。

## 4. 生产环境替换

保持生产 PostgreSQL 和 Redis 完全不变。只修改 New API 应用镜像标签或应用代码路径。

如果你要保持原版经典前端，启动后请确认前端主题配置为 `classic`。不要把 `theme.frontend` 切换为 `default`，除非你明确想使用新版前端。

推荐的镜像标签流程：

```bash
cd /opt/new-api
docker compose pull new-api
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

如果你在服务器本地构建：

注意区分两个目录：

- `/opt/newapi` 或 `/opt/new-api`：生产部署目录，通常只有 `docker-compose.yml` 和 `.env`，没有 `.git` 是正常的。
- `/opt/src/newapi0604`：源码目录，用于从 GitHub 拉取二开版本并构建镜像。

不要在生产部署目录里执行 `git remote`、`git fetch`、`git checkout`，除非该目录本身就是源码仓库。

先创建或进入源码目录。如果源码目录已经存在，检查它当前的 GitHub 源；如果仍然指向旧仓库，需要先切换到新的仓库地址：

```bash
mkdir -p /opt/src
cd /opt/src

if [ ! -d newapi0604/.git ]; then
  git clone https://github.com/zhangdc1/newapi0604.git newapi0604
fi

cd /opt/src/newapi0604
git remote -v
git branch --show-current
git status --short
```

如果 `origin` 不是 `https://github.com/zhangdc1/newapi0604.git`，先记录旧地址，再切换：

```bash
git remote get-url origin > backup.git-origin.$(date +%F-%H%M%S).txt
git remote set-url origin https://github.com/zhangdc1/newapi0604.git
git remote -v
```

然后拉取新仓库代码。以下命令默认使用 `main` 分支；如果新仓库实际分支不是 `main`，先用 `git branch -r` 查看后替换分支名：

```bash
git fetch origin main
git checkout -B main origin/main
docker build -t new-api:redevelopment-0604 .
```

然后回到生产部署目录，只更新 `docker-compose.yml` 中 New API 服务的镜像：

```yaml
services:
  new-api:
    image: new-api:redevelopment-0604
```

只应用应用服务：

```bash
docker compose up -d --no-deps new-api
```

不要运行：

```bash
docker compose down -v
docker volume rm <any_production_volume>
```

## 5. 升级后检查

立即验证：

```bash
docker compose ps
docker compose logs --tail=200 new-api
```

应用检查清单：

- 管理员登录正常。
- 用户登录正常。
- 用户余额、令牌、渠道和设置都还在。
- 充值页面仍然显示现有支付方式。
- 兑换码流程仍然正常。
- 用户可以打开分销中心。
- 管理员可以打开分销管理。
- 新邀请用户通过 `?aff=<code>` 注册时，只绑定一次邀请人。
- 奖励结算只会把佣金转入余额，不会再次触发佣金。

## 6. 回滚

升级前保留旧镜像标签：

```bash
docker image ls | grep new-api
```

如果新应用不健康：

```bash
cd /opt/new-api
```

将 `docker-compose.yml` 中 New API 服务的镜像改回之前的标签，然后执行：

```bash
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

本次重构会新增独立的分销表，并保持旧表兼容。旧版本应该会忽略这些新表。回滚期间避免手动清理数据库结构。

## 7. 数据安全说明

此版本使用追加式数据库变更：

- 新增 `distribution_settings`。
- 新增 `distribution_commission_records`。
- 新增 `distribution_transfers`。
- 在可用时复用现有用户邀请字段，例如 `aff_code`、`inviter_id`、`aff_quota` 和 `aff_history`。

任何迁移都不应该删除生产表、清空数据、重建用户、重建配置或重置余额。

如果某个部署脚本建议删除卷、初始化空数据库或替换 `.env`，请停止并先审查。
