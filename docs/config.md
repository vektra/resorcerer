Resorcerer uses YAML for configuration for uniformity.

The system works by using an internal event system where service definitions emit event that the system routes to handlers to perform an action.

# Toplevel Keys

- **mode** - Valid values: upstart
- **poll**
  - _seconds_ - Default: 5. How many seconds to poll for serice information
  - _samples_ - Default: 5. The window of samples to keep to calculate if a metric has been overrun
  - _significant_ - Default: (_samples_ / 2) + 1. The number of samples in the window over the limit to trip the event
- **email**
  - _server_ - host:port of server to send email through
  - _username_ - username to authenticate with
  - _password_ - password to authenticate with
- **services** - An array of service definitions. See Services below
- **on** - An array of handler definitons that apply to all services. See Handlers below

## Services

A service is configured as a map with the following keys:

- **name** - The name of the service as given by your process management framework (right now, only Upstart)
- **memory** - The memory limit on the process. Value is in bytes unless the suffix kb for kilobytes or mb for megabytes is added.
- **on** - An array of handler definitions that only apply to this service. See Handlers below

## Handlers

Handlers attach an action to an event. A handler is configured with the following keys:

- **event** - The name of the event. See Events below
- **process** - Optional. Change the services status. Possible values: restart, stop.
- **email** - Send an email about the event, configured with these keys:
  - _address_ - The email address to send to
  - _subject_ - Optional. A prefix for the subject line of the email
- **script** - Value: A script to run using bash. The stdin to the command is a JSON representation of the event
- **webhook** - Value: A URL to POST to. The body of the request is a JSON representation of the event


## Events

This is a list of all events that handlers can bind to. The events are organized in hierarchies and the handlers are searched within the heirarchy.

- **memory/measured** - Emitted each time the memory for a service is measured
- **memory/limit/over** - Emitted when a services memory is over the limit a significant number of times
- **memory/limit/recover** - Emitted when a services memory has returned to normal
- **monitor/start** - Emitted when resorcerer is loading, once per service
- **monitor/down** - Emitted when a service is detected as not running
- **monitor/up** - Emitted when a service that was down is now running
- **monitor/pid-change** - Emitted when a service's pid changes
- **action/error** - Emitted when a handlers action errors out. The values is: (original\_event, error)
