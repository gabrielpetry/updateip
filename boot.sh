#!/bin/bash

pip="pip3"
python="python3"

if python --version | grep -q " 3."; then
    pip="pip"
    python="python"
fi

cd "$(dirname $0)" || exit 1

if ! test -f .env; then
    echo " NO ENV FILE "
    exit 1
fi

git pull origin master

if test -d venv; then
    source venv/bin/activate
else
    if command -v virtualenv >/dev/null 2>&1; then
        virtualenv -p "$python" venv
        source venv/bin/activate
    fi
fi

$pip install -r requirements.txt

$python update_ip.py

if crontab -l | grep -q update_ip; then
    echo "Creating cron job"
    (crontab -l 2>/dev/null || true; echo "*/5 * * * * $PWD/boot.sh") | crontab -
fi