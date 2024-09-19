FROM nginx:latest
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 4040
COPY creds /creds
CMD ["/bin/sh", "-c", " sed -i \"s/<SECRET_CREDS>/$(cat /creds)/\" /etc/nginx/conf.d/default.conf && nginx -g \"daemon off;\""]