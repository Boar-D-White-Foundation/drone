#!/usr/bin/env bash
set -x

cd /home/fh/dev/drone || exit
docker-compose down
sudo tar --owner=fh --group=fh -czvf data_backup.tar.gz data/
# rclone config -> add ydrive & gdrive remotes
rclone sync -v data_backup.tar.gz "ydrive:drone/backup_$(date +%a)"
rclone sync -v data_backup.tar.gz "gdrive:drone/backup_$(date +%a)"
docker-compose up -d
