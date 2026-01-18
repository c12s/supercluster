#!/bin/sh

kraft run \
  --rm \
  -p 8002:8002 \
  --plat qemu \
  --arch x86_64 \
  .

