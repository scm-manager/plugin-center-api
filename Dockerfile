FROM scratch
COPY target/plugin-center-api /plugin-center-api
COPY plugin-center/src/plugins /plugins
USER 10000
ENV CONFIG_DESCRIPTOR_DIRECTORY /plugins
EXPOSE 8000
CMD ["/plugin-center-api"]
