FROM node:22-alpine AS build

WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM nginx:1.29-alpine

# главный конфиг: рабочий deploy/nginx.conf, если он есть, иначе шаблон
COPY deploy/nginx.example.conf deploy/nginx.con[f] /tmp/nginx/
RUN cp /tmp/nginx/nginx.conf /etc/nginx/nginx.conf 2>/dev/null || cp /tmp/nginx/nginx.example.conf /etc/nginx/nginx.conf; \
    rm -rf /tmp/nginx

COPY deploy/nginx.site.conf /etc/nginx/conf.d/default.conf
COPY --from=build /web/dist /usr/share/nginx/html
