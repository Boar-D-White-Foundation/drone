#!/usr/bin/env bash
set -euxo pipefail

cd /home/fh/dev/drone || exit
# stop drone before backup
docker compose down

# create backups
# sudo is needed to access files written by root inside docker
sudo tar -czvf data_backup.tar.gz data/
sudo go run ./cmd/db-dump -p db_dump.json.gz
sudo chown fh:fh data_backup.tar.gz db_dump.json.gz

# test db restore
sudo rm -rf data/badger
sudo go run ./cmd/db-restore -p db_dump.json.gz

# upload backups to cloud
# to work properly needs: rclone config -> add ydrive & gdrive remotes
rclone sync -v data_backup.tar.gz "ydrive:drone/backup_$(date +%a)"
rclone sync -v db_dump.json.gz "ydrive:drone/backup_$(date +%a)"
rclone sync -v data_backup.tar.gz "gdrive:drone/backup_$(date +%a)"
rclone sync -v db_dump.json.gz "gdrive:drone/backup_$(date +%a)"

# start drone after backup
docker compose up -d
