# New API Production-Safe Upgrade Guide

This guide is for replacing an existing production New API service with this rebuilt version while keeping PostgreSQL, Redis, mounted data, users, balances, tokens, channels, recharge records, redemption records, and system options.

The core rule is simple: replace only the application image or application code. Do not replace the production database, Redis volume, `.env`, or Docker volumes.

## 1. Production Asset Inventory

Before touching production, record the current values:

```bash
cd /opt/new-api
docker compose ps
docker compose config > backup.compose.rendered.$(date +%F-%H%M%S).yml
```

Inventory checklist:

- Current `docker-compose.yml` path.
- Current `.env` path.
- New API service name, image tag, container name, and mounted data directory.
- PostgreSQL service/container name, database name, username, password source, and volume name.
- Redis service/container name and volume name.
- Reverse proxy config, such as Nginx/Caddy/Cloudflare tunnel.
- Current production image tag, for rollback.

Useful commands:

```bash
docker volume ls
docker inspect <postgres_container> --format '{{json .Mounts}}'
docker inspect <redis_container> --format '{{json .Mounts}}'
docker inspect <newapi_container> --format '{{json .Mounts}}'
```

## 2. Backup Before Upgrade

Create a timestamp and backup directory:

```bash
TS=$(date +%F-%H%M%S)
mkdir -p /opt/new-api-backups/$TS
cd /opt/new-api
```

Back up PostgreSQL:

```bash
docker exec <postgres_container> pg_dump \
  -U <postgres_user> \
  -d <postgres_database> \
  -Fc \
  -f /tmp/newapi-$TS.dump

docker cp <postgres_container>:/tmp/newapi-$TS.dump /opt/new-api-backups/$TS/
```

Back up Redis persistence or volume data. If Redis uses an appendonly file or `dump.rdb` in a mounted directory:

```bash
cp -a /path/to/redis/data /opt/new-api-backups/$TS/redis-data
```

If Redis uses a Docker volume:

```bash
docker run --rm \
  -v <redis_volume>:/data:ro \
  -v /opt/new-api-backups/$TS:/backup \
  alpine tar czf /backup/redis-volume.tgz -C /data .
```

Back up compose, env, and application mounted data:

```bash
cp -a docker-compose.yml /opt/new-api-backups/$TS/
cp -a .env /opt/new-api-backups/$TS/
cp -a /path/to/new-api/data /opt/new-api-backups/$TS/new-api-data
```

Forbidden during backup and upgrade:

- Do not run `docker compose down -v`.
- Do not delete Docker volumes.
- Do not overwrite production `.env`.
- Do not initialize an empty production database.

## 3. Staging Verification With Production Backup

Restore the production PostgreSQL backup into a test database or test server first:

```bash
createdb newapi_upgrade_test
pg_restore -U <postgres_user> -d newapi_upgrade_test /opt/new-api-backups/$TS/newapi-$TS.dump
```

Start the new application against the test database and a test Redis instance. Use copied `.env` values, but change only:

- PostgreSQL host/database to the test database.
- Redis host/database to the test Redis.
- Public callback/domain values if testing payment callbacks.

Then start the new image:

```bash
docker compose -f docker-compose.upgrade-test.yml up -d
docker compose -f docker-compose.upgrade-test.yml logs -f new-api
```

Verification checklist:

- Existing users can log in.
- Existing balances are unchanged.
- Existing API tokens still work.
- Existing channels, model ratios, groups, and options remain.
- Existing recharge and redemption histories remain.
- Existing Redis-backed sessions/cache do not cause startup errors.
- New distribution tables are created automatically.
- Distribution system is disabled by default until an admin enables it.
- DeepSeek calls record official `usage` when upstream returns it.

If any required item fails in staging, do not touch production.

## 4. Production Replacement

Keep production PostgreSQL and Redis exactly as they are. Change only the New API app image tag or application code path.

Recommended image-tag flow:

```bash
cd /opt/new-api
docker compose pull new-api
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

If you build locally on the server:

```bash
git fetch origin main
git checkout main
docker build -t new-api:redevelopment-0604 .
```

Then update only the New API service image in `docker-compose.yml`:

```yaml
services:
  new-api:
    image: new-api:redevelopment-0604
```

Apply only the application service:

```bash
docker compose up -d --no-deps new-api
```

Do not run:

```bash
docker compose down -v
docker volume rm <any_production_volume>
```

## 5. Post-Upgrade Checks

Immediately verify:

```bash
docker compose ps
docker compose logs --tail=200 new-api
```

Application checklist:

- Admin login works.
- User login works.
- User balance, tokens, channels, and settings are present.
- Recharge page still shows existing payment methods.
- Redemption code flow still works.
- Distribution Center loads for a user.
- Distribution Management loads for an admin.
- New invited user registration through `?aff=<code>` binds inviter once.
- Reward settlement transfers only commission to balance and does not trigger another commission.

## 6. Rollback

Keep the old image tag before upgrade:

```bash
docker image ls | grep new-api
```

If the new application is unhealthy:

```bash
cd /opt/new-api
```

Set the New API service image back to the previous tag in `docker-compose.yml`, then:

```bash
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

This redevelopment adds separate distribution tables and keeps old tables compatible. The old version should ignore the new tables. Avoid manual schema cleanup during rollback.

## 7. Data Safety Notes

This version uses additive database changes:

- Adds `distribution_settings`.
- Adds `distribution_commission_records`.
- Adds `distribution_transfers`.
- Reuses existing user invitation fields where available, such as `aff_code`, `inviter_id`, `aff_quota`, and `aff_history`.

No migration should drop production tables, truncate data, recreate users, recreate options, or reset balances.

If a deployment script suggests deleting volumes, initializing an empty database, or replacing `.env`, stop and review before proceeding.
