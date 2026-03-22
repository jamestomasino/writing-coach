# Staging And Rollback

This project stores production state in Docker named volumes, not host files.

That means the safe pre-deploy path is:

1. back up the production volumes
2. copy those volumes into a separate staging project
3. start the staging stack on a different localhost port
4. verify the app against realistic data
5. only then rebuild the production stack

## Files

- Staging Compose override: [docker-compose.staging.yml](/home/tomasino/writing-coach/docker-compose.staging.yml)
- Backup script: [backup-compose-volumes.sh](/home/tomasino/writing-coach/scripts/backup-compose-volumes.sh)
- Staging copy script: [prepare-staging-volumes.sh](/home/tomasino/writing-coach/scripts/prepare-staging-volumes.sh)
- Restore script: [restore-compose-volumes.sh](/home/tomasino/writing-coach/scripts/restore-compose-volumes.sh)

## One-Time Server Setup

Make the scripts executable:

```bash
chmod +x scripts/backup-compose-volumes.sh
chmod +x scripts/prepare-staging-volumes.sh
chmod +x scripts/restore-compose-volumes.sh
```

The staging stack defaults to:

- Compose project: `writing-coach-staging`
- Web bind: `127.0.0.1:12234`
- Public URL assumption: `http://localhost:12234`

In the copied staging checkout, create a staging env file from the live one and change the bind/public URL values there:

```bash
cp .env .env.staging
sed -i 's#^WEB_PORT_BIND=.*#WEB_PORT_BIND=127.0.0.1:12234:3000#' .env.staging
sed -i 's#^COACH_PUBLIC_URL=.*#COACH_PUBLIC_URL=http://localhost:12234#' .env.staging
```

If you want to browse the staging stack remotely, use an SSH tunnel:

```bash
ssh -L 12234:127.0.0.1:12234 your-server
```

Then open `http://localhost:12234`.

## Normal Pre-Deploy

From the repo on the server:

```bash
git pull --ff-only
./scripts/backup-compose-volumes.sh
docker compose -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml down
./scripts/prepare-staging-volumes.sh writing-coach writing-coach-staging
cp .env .env.staging
sed -i 's#^WEB_PORT_BIND=.*#WEB_PORT_BIND=127.0.0.1:12234:3000#' .env.staging
sed -i 's#^COACH_PUBLIC_URL=.*#COACH_PUBLIC_URL=http://localhost:12234#' .env.staging
docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml up -d --build
```

Then verify:

```bash
docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml ps
docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml logs --tail=200 app
curl -I http://127.0.0.1:12234
```

Recommended manual checks in staging:

- open the app and confirm it loads
- confirm existing users still have expected tracks
- inspect a few active TGOs and completed TGOs
- open a recent assignment and review
- test login and a normal practice flow if the release touches auth or session state

When done:

```bash
docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml down
```

## Production Deploy

If staging looks good, deploy production normally:

```bash
docker compose up -d --build
docker compose ps
docker compose logs --tail=200 app
curl -I http://127.0.0.1:11234
```

## Rollback

If the deploy is bad, roll back both code and volumes.

1. Stop the broken stack:

```bash
docker compose down
```

2. Check out the last known good revision:

```bash
git log --oneline -n 10
git checkout <known-good-commit>
```

3. Restore the production backup you made before deploy:

```bash
./scripts/restore-compose-volumes.sh ./backups/<backup-dir>
```

4. Bring the old revision back up:

```bash
docker compose up -d --build
docker compose ps
docker compose logs --tail=200 app
```

## Practical Habit

The steady-state release loop should now be:

1. `git pull --ff-only`
2. `./scripts/backup-compose-volumes.sh`
3. `./scripts/prepare-staging-volumes.sh writing-coach writing-coach-staging`
4. create `.env.staging` with `WEB_PORT_BIND=127.0.0.1:12234:3000` and `COACH_PUBLIC_URL=http://localhost:12234`
5. `docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml up -d --build`
6. verify staging
7. `docker compose --env-file .env.staging -p writing-coach-staging -f docker-compose.yml -f docker-compose.staging.yml down`
8. `docker compose up -d --build`

That keeps the migration and application startup on realistic copied data before production sees the release.
