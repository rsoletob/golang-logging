Logging library that uses [logrus](http://github.com/Sirupsen/logrus) to write json lines with more standard fields for strutured logging.

To enable debug logging set env variable ```DEBUG``` to ```1```.

To disable json-line format (e.g. to get get human-readable lines for local testing) set env variable ```LOG_FORMAT``` to ```plain```.