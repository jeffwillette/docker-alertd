# Docker-Alertd

[![Build Status](https://travis-ci.org/deltaskelta/docker-alertd.svg?branch=master)](https://travis-ci.org/deltaskelta/docker-alertd)

## What Does It Do?

docker-alertd monitors docker containers on a host machine and sends alerts via email when usage limits (as defined in a conf file) have been breached. It can be run in the foreground or background

Current metrics that can be tested are:

1. Memory usage (in MB)
2. CPU Usage (as a percentage)
3. Minimum Process running in container

# Step 1: Install

### Method 1: Download a pre-compiled binary

####[linux](https://jrwillette.com/media/binaries/linux/docker-alertd), [macOS](https://jrwillette.com/media/binaries/macOS/docker-alertd), [windows](https://jrwillette.com/media/binaries/windows/docker-alertd.exe)

### Method 2: Build from source

Assuming that you already have `go` installed on your machine, you can build from
source...

```
go get github.com/deltaskelta/docker-alertd
cd $GOPATH/src/github.com/deltaskelta/docker-alertd
go install
```

`go install` will compile and install the binary to you users `$GOPATH`. More information about how to properly setup a go environment can be found [at the go website](https://golang.org/doc/install)

# Step 2: Make a Configuration File

Docker-Alertd takes one argument which is the path to a configurations file. The configuration file format is in JSON format, it consists of one object, which should include an array of at least one container, and valid email credentials to login and send mail.

#####Example `conf.json` file
```json
{
	"containers": [
		{
			"name": "container1",
			"max-cpu": 0,
			"max-mem": 20,
			"min-procs": 3
		},
		{
			"name": "container2",
			"max-cpu": 20,
			"max-mem": 20,
			"min-procs": 4
		}
	],
	"email_settings": {
		"from": "auto@freshpowpow.com",
		"to": "jeff@gnarfresh.com",
		"smtp": "smtp.coolserver.com",
		"password": "gnarlesbarkely",
		"port": 587
	}
}
```

### Configuration Variables

##### Containers
1. name: the container name or ID
2. max-cpu: the maximum cpu usage threshold (as a percentage), if the container uses more CPU, an alert will be triggered.
3. max-mem: the maximum memory usage threshold (in MB). If the container uses more system memory than this, an alert will be triggered.
4. min-procs: the minimum number of running processes (PID's) in the container. If a the number of running processes dips below this level (when a process fails), an alert will be triggered.

##### Email Settings
1. from: the email address to send alert emails from
2. to: the email address to send alert emails to (can be the same as from)
3. smtp: the smtp server to connect to
4. password: the password to use for smtp authentication
5. port: the port to connect to the smtp server

# Step 3: Run the program

The program has one required option ( -f [config-file]) and needs to be started with the path to the configuration file

```
/path/to/binary/docker-alertd -f ~/path/to/configuration/file/config.json
```

This will start the program and log the output to stdout. It can be stopped with CTRL-C.

#### Example Output:

```
2017/02/17 11:46:40 started docker-alertd process
------------------------------
2017/02/17 11:46:42 CPU ALERT: container1's CPU usage exceeded 0, it is currently using 0.101465
2017/02/17 11:46:47 alert email sent
```

# 4 Set up as a background process (optional)

If you wish to have docker-alertd run as a background process, it needs to be setup as a background process as per your operating system.

#### As A Systemd Service (for Linux systems with systemd)

If you have a systemd based system then you can refer to [docker-alertd.service.example](https://github.com/deltaskelta/docker-alertd/blob/master/docker-alertd.service.example) the example systemd service file and this [tutorial](https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units)

#### With Launchd (MacOS)

Refer to the [launchd plist example file](https://github.com/deltaskelta/docker-alertd/blob/master/com.github.docker-alertd.plist.example) file and the [launchd reference](http://www.launchd.info/)

#### With Sys V Init (various Linux systems without systemd)

Refer to this [Sys V Init tutorial](https://www.cyberciti.biz/tips/linux-write-sys-v-init-script-to-start-stop-service.html)

