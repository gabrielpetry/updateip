#!/bin/bash


cd "$(readlink -f \"$0\")" || exit 1

git pull origin master

source venv/bin/activate

pip install -r requirements.txt

python update_ip.py
