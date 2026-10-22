#!/bin/sh

kraft run \
  --rm \
  -p 8000:8000 \
  --plat qemu \
  --arch x86_64 \
  .

