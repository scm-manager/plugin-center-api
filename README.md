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
