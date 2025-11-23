FROM debian:stable-slim

ARG TARGETARCH
ENV ARCH=${TARGETARCH}

COPY bin/discord-werewolf-linux-${ARCH} /usr/local/bin/discord-werewolf

RUN useradd app && \
    mkdir /data && \
    chown app:app /data && \
    chmod +x /usr/local/bin/discord-werewolf

USER app
ENTRYPOINT ["/usr/local/bin/discord-werewolf"]
CMD ["/data/discord-werewolf.db"]