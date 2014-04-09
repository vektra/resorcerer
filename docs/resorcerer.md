# Resorcerer

Resorcerer is a resource monitoring and reaction tool. It provides the ability to set limits on the resource that a service can use then wire up ways to react to breakages in those limits.

## Usage

> $ resorcerer [--debug] [--dryrun] config.yml

The __--debug__ option turns on additional output while the tool runs, allowing you to follow along at home.

__--dryrun__ is similar to __--debug__ but it disables running the actions, instead the system just indicates that they would have run. This is good for debugging a new config.

## Config

See config.md for details about the format of the config file

## Signals

Resorcerer responds to SIGHUP to reload it's configuration, allowing for nice integration with config management tools.


