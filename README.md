# divido-cli
A CLI to help divido devs

# Definitions

- Envrionment is like ing-test, it has a helm chart and overrides
- Charts is a helm chart, it has a list of services
- Service like portals-web-pub have a version

# List of things you can do 

- show services deployed in an environment
- show services in a helm chart
- diff between helm charts
- generate changelog between two given helm charts (v1.2.3 -> v1.2.4)
- bump a service in a helm chart
- bump a helm chart in an environment 
- override/remove override a service in an environment
- undo / redo last commands
- set GITHUB_TOKEN

All via GITHUB API, grabbed GITHUB_TOKEN from env variable if possible. Maybe also an option to set it 
