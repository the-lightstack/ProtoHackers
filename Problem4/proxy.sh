#!/bin/bash
socat -v udp-listen:13337,fork,reuseaddr "udp:[::]:1337"
