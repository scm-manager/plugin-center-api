FROM alpine:3.10.3
RUN apk --no-cache add ca-certificates
COPY target/plugin-center-api /plugin-center-api
COPY website/content/plugins /plugins
COPY website/content/plugin-sets /plugin-sets
USER 10000
ENV CONFIG_DESCRIPTOR_DIRECTORY /plugins
ENV CONFIG_PLUGIN_SETS_DIRECTORY /plugin-sets
EXPOSE 8000
CMD ["/plugin-center-api"]
