FROM crawlabteam/crawlab-backend:latest AS backend-build

FROM crawlabteam/crawlab-frontend:latest AS frontend-build

FROM crawlabteam/crawlab-base:latest

# copy backend files
RUN mkdir -p /opt/bin
COPY --from=backend-build /go/bin/crawlab /opt/bin
RUN cp /opt/bin/crawlab /usr/local/bin/crawlab-server

# copy backend config files
COPY ./backend/conf /app/backend/conf

# copy frontend files
COPY --from=frontend-build /app/dist /app/dist

# copy nginx config files
COPY ./docker/nginx/crawlab.conf /etc/nginx/conf.d

# copy docker bin files
COPY ./docker/bin /app/

# start backend
CMD ["/bin/bash", "/app/bin/docker-init.sh"]
