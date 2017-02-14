# docker-alertd

## Compiled Binaries

####[linux](https://jrwillette.com/media/binaries/linux/docker-alertd)
####[macOS](https://jrwillette.com/media/binaries/macOS/docker-alertd)
####[windows](https://jrwillette.com/media/binaries/windows/docker-alertd.exe)

## Intro

docker-alertd monitors docker containers on a machine and sends alerts visa email when usage limits have been breached.

Current metrics that can be tested are:

1. Memory usage (in MB)
2. CPU Usage (as a percentage)
3. Minimum Process running in container

## Config File

The configuration file format is in JSON format, it need to include an array of at least one container, and email settings to login and send mail. 

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

## Running the program

The program has one required option (config file) and needs to be started with the path to the configuration file

```
docker-alertd -f ~/path/to/configuration/file/config.json
```
