FROM nginx:1.27-alpine

RUN apk add --no-cache brotli

COPY nginx.conf /etc/nginx/nginx.conf
COPY . /usr/share/nginx/html

RUN find /usr/share/nginx/html -type f \( -name "*.html" -o -name "*.css" -o -name "*.js" -o -name "*.json" -o -name "*.svg" \) \
    -exec gzip -k -f -9 {} \; \
    -exec brotli -k -f -q 11 {} \;

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
