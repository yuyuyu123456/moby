package daemon

import (
	// Importing packages here only to make sure their init gets called and
	// therefore they register themselves to the logdriver factory.
	_ "moby/daemon/logger/awslogs"
	_ "moby/daemon/logger/fluentd"
	_ "moby/daemon/logger/gcplogs"
	_ "moby/daemon/logger/gelf"
	_ "moby/daemon/logger/journald"
	_ "moby/daemon/logger/jsonfilelog"
	_ "moby/daemon/logger/logentries"
	_ "moby/daemon/logger/splunk"
	_ "moby/daemon/logger/syslog"
)
