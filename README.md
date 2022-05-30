# Plugin Center API

Rest API for the SCM-Manager Plugin Center

## Configuration

The application can be configured with an yaml file or over environment variables.
The location of the configuration can be specified with the `CONFIG` environment variable, 
if nothing is specified `config.yaml` is used.
The following parameters can be configured:

| Yaml Key              | Environment Variable         | Default value |
|-----------------------|------------------------------|---|
| descriptor-directory  | CONFIG_DESCRIPTOR_DIRECTORY  | - |
| plugin-sets-directory | CONFIG_PLUGIN_SETS_DIRECTORY | - |
| port                  | CONFIG_PORT                  | 8000 |

## Test locally

1. Build executable:
   ```
   go build -a -tags netgo -ldflags "-w -extldflags \'-static\'" -o target/plugin-center-api *.go
   ```
2. Clone `website` repository directly into project's root folder
3. Build docker image:
   ```
   docker build -t scmmanager/plugin-center-api .
   ```
4. Run plugin-center-api:
   ```
   docker run -p 8000:8000 -d scmmanager/plugin-center-api
   ```
