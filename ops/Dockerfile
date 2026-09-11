# Self-contained arm64 build for the on-server builder (devops/build/build.sh).
# Context is the repo root: the site imports ../examples/*/main.go at build
# time, so both site/ and examples/ must be in the image's build stage.

# --- 1. Astro static build -> /src/site/dist ---------------------------------
FROM node:22-alpine AS build
WORKDIR /src/site
COPY site/package.json site/package-lock.json ./
RUN npm ci
COPY examples/ /src/examples/
COPY docs/logo.svg /src/docs/logo.svg
COPY site/ ./
RUN npm run build

# --- 2. nginx runtime ---------------------------------------------------------
FROM nginx:alpine
COPY ops/nginx.conf /etc/nginx/nginx.conf
COPY --from=build /src/site/dist/ /usr/share/nginx/html/
EXPOSE 8080
USER nginx
