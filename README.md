# goodhost

> hosts, for good

### Install
For now, only via the Go toolchain:
```bash
go install github.com/pbdeuchler/goodhost
```
- Tested on OSX 10.11 (All others YMMV)
- Not for production use

### Commands
> Note: All commands omit the default entries from removal and display

#### list
Display a formatted list of all entries.

```bash
23:28:15: /Users/foodbuttondev/workspace/gocode/src/github.com/foodbutton/burritobutton (master*) 
 → goodhost list
+-----------------+---------------------------+------------+
| NETWORK ADDRESS |         HOSTNAME          |   LABEL    |
+-----------------+---------------------------+------------+
| 245.81.229.66   | www.jenkins.foodbutton.io |            |
| 61.61.113.198   | weee.foodbutton.io        | haproxy LB |
| 126.48.211.25   | lb.foodbutton.io          | LB         |
| 24.174.76.168   | web.foodbutton.io         | staging    |
| 112.101.132.131 | admin.foodbutton.io       |            |
| 119.104.50.216  | api.foodbutton.io         |            |
| 51.65.136.224   | worker.foodbutton.io      |            |
| 168.229.143.129 | apilb.foodbutton.io       | LB         |
| 67.211.203.212  | splunk.foodbutton.io      | splunk     |
| 192.168.99.100  | local.dev                 |            |
+-----------------+---------------------------+------------+
```

#### get
Query entries by network address.
```bash
23:28:15: /Users/foodbuttondev/workspace/gocode/src/github.com/foodbutton/burritobutton (master*) 
 → goodhost get 67.211.203.212
+-----------------+----------------------+--------+
| NETWORK ADDRESS |       HOSTNAME       | LABEL  |
+-----------------+----------------------+--------+
| 67.211.203.212  | splunk.foodbutton.io | splunk |
+-----------------+----------------------+--------+
```

#### set
Add an entry with an optional label (labels are added as comments inline).
```bash
23:28:15: /Users/foodbuttondev/workspace/gocode/src/github.com/foodbutton/burritobutton (master*) 
 → goodhost set 67.211.203.213 grafana.foodbutton.io "grafana"
```

```bash
# /etc/hosts
...
67.211.203.213	grafana.foodbutton.io	# grafana
```

#### remove
Remove an entry by network address.
```bash
23:28:15: /Users/foodbuttondev/workspace/gocode/src/github.com/foodbutton/burritobutton (master*) 
 → goodhost remove 67.211.203.213
```

#### label
Add or replace the label of an existing entry.
```bash
23:28:15: /Users/foodbuttondev/workspace/gocode/src/github.com/foodbutton/burritobutton (master*) 
 → goodhost label 119.104.50.216 "api v2"
```

```bash
# /etc/hosts
...
119.104.50.216	api.foodbutton.io	# api v2
```

### Addendum
Nervous? Testing? Have a workflow you're ashamed of? By default `goodhost` messes with `/etc/hosts`, but you can use the `--file, -f` flag or set your `$HOSTS_FILE` environment variable to change what file `goodhost` manipulates.

### ToDo
- Flag to output list in JSON format
- Get by label
- Validate network addresses