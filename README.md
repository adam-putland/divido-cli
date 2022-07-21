
# divido-cli

An interactive prompt cli to help Divido devs manage services and GitHub Helm updates

All via GITHUB API, grabbed GITHUB_TOKEN from env variable if possible.

## Usage
In order to utilize this you're going to need:
- A GitHub Personal Access token setup and configured on your GitHub profile and local machine

    - You can setup a token by logged into GitHub, [visit your account tokens page](https://github.com/settings/tokens) visit your account tokens page —> create a new token
    - The only permissions required is `repo`
    - Open a new terminal and run `export GITHUB_TOKEN="[TOKEN HERE]"` (no square brackets or quotes)
    - run `echo $GITHUB_TOKEN` —> verify you see your token
  
## Features

- show services deployed in an environment  (e.g. see all services deployed in ING testing)
- show services in a helm chart  (e.g. see all services in a specific (v1.31.65) ING Helm chart)
- diff between helm charts 
- generate changelog between two given helm charts (v1.2.3 -> v1.2.4)
- update a service version in a helm chart
- update a helm chart version in an environment

### Show Service information (e.g portals-web-pub)

![Gif](./assets/services.gif)


### Show services deployed in a specific helm (e.g. ING)

![Gif](./assets/helm-services-info.gif)

### Compare services deployed in a specific helm versions and export release (e.g ING)

![Gif](./assets/helm-export-release.gif)

### Update services deployed in a specific helm (e.g. ING)
![Gif](./assets/helm-update-services.gif)


### Show services in a specific env and update its helm version  (e.g. ING test)
![Gif](./assets/env-show-and-update-version.gif)

