FROM scratch
COPY target/plugin-center-api /plugin-center-api
USER 10000
CMD ["/plugin-center-api"]
