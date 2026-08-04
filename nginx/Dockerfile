# syntax=docker/dockerfile:1

# Official nginx, pinned to a specific minor.
FROM nginx:1.27

# Replace the image's default "Welcome to nginx" server block with custom reverse-proxy config.
# /etc/nginx/conf.d/default.conf is the path the stock image includes by default
COPY nginx.conf /etc/nginx/conf.d/default.conf