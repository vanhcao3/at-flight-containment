ARG BUILDER_IMAGE=harbor.vht.vn/c4i/golang:1.24.0-bookworm
ARG TARGET_IMAGE=scratch

FROM ${BUILDER_IMAGE} AS builder

WORKDIR /at-drone

COPY . .

RUN make go-build


FROM ${TARGET_IMAGE} AS final

WORKDIR /at-drone

COPY --from=builder /at-drone/bin/at-drone ./bin/at-drone
COPY --from=builder /at-drone/web ./bin
COPY --from=builder /at-drone/etc/app.yaml ./etc/app.yaml
COPY --from=builder /at-drone/data ./data

ENTRYPOINT ["./bin/at-drone", "start"]

CMD ["./etc"]
