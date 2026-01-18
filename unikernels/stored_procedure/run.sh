#!/bin/sh

kraft run \
  --rm \
  -p 8080:8080 \
  --plat qemu \
  --arch x86_64 \
  .

