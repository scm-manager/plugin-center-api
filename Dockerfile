FROM alpine:3.10.3
RUN apk --no-cache add ca-certificates
COPY target/plugin-center-api /plugin-center-api
COPY website/content/plugins /plugins
USER 10000
ENV CONFIG_DESCRIPTOR_DIRECTORY /plugins
EXPOSE 8000
CMD ["/plugin-center-api"]
