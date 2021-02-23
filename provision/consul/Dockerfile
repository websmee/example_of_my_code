# pull official base image
FROM consul

# key-value shim loader
# you can override this loader using an empty docker volume:
ENV INIT_CONSUL_KV_DIR=/var/local/init_consul_kv.d
RUN mkdir -p $INIT_CONSUL_KV_DIR
COPY kv_loader.sh /usr/local/bin/
RUN sed -ie 's|^\(set .*\)|\1\n/usr/local/bin/kv_loader.sh \&|' /usr/local/bin/docker-entrypoint.sh
COPY init_kv $INIT_CONSUL_KV_DIR/