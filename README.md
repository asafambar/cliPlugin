# Release-Notes

## About this plugin
This plugin retrieve and present Jfrog products Release Notes from JFrog CLI.
Support Release notes of the current default Jfrog CLI products or specific product and version.

## Installation with JFrog CLI
Since this plugin is currently not included in [JFrog CLI Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg), it needs to be built and installed manually. Follow these steps to install and use this plugin with JFrog CLI.
1. Make sure JFrog CLI is installed on you machine by running ```jfrog```. If it is not installed, [install](https://jfrog.com/getcli/) it.
2. Create a directory named ```plugins``` under ```~/.jfrog/``` if it does not exist already.
3. Clone this repository.
4. CD into the root directory of the cloned project.
5. Run ```go build``` to create the binary in the current directory.
6. Copy the binary into the ```~/.jfrog/plugins``` directory.

## Usage
### Commands
* release-notes(alias: rn)
    - Arguments:
        - product - The name of the product you would like to get release notes.
        - version - The version of product you would like to get release notes.
    - Flags:
        - current: Get release notes for current Jfrog CLI default specific product.
        - date: Get only the date of the release for specific product and version.
    - Example:
    ```
  $ jfrog rn -current xray
  $ jfrog rn artifactory 7.11.5
  
  result1:
  ## Xray 3.12
  
  Released: November 29, 2020
  
  #### Feature Enhancements
   ......
  
  rsult2:

  ### Artifactory 7.11.5
  
  Released: 1 December 2020
  
  #### Resolved Issues
   ......
  ```

## Additional info
Currently support Artifactory and Xray.

## Release Notes
The release notes are available [here](RELEASE.md).
