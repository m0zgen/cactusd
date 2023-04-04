[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/m0zgen/cactusd/release.yml "Release")](https://github.com/m0zgen/cactusd/actions/workflows/release.yml)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m0zgen/cactusd "Go version")](#)
[![GitHub Release Date](https://img.shields.io/github/release-date/m0zgen/cactusd "Latest release date")](https://github.com/m0zgen/cactusd/releases)
[![GitHub latest version](https://img.shields.io/github/v/release/m0zgen/cactusd "Latest version")](https://github.com/m0zgen/cactusd/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/m0zgen/cactusd)](https://goreportcard.com/report/github.com/m0zgen/cactusd)
[![CodeQL](https://github.com/m0zgen/cactusd/actions/workflows/codeql.yml/badge.svg?branch=dev&event=push)](https://github.com/m0zgen/cactusd/actions/workflows/codeql.yml)

# CACTUSD (Command and Actions Routine Server Daemon)

Download/Upload, merge, divide by category and publish a domain name block lists and IP addresses as merged files.

* Block and Allow lists periodically updated (every 12 hours), you can download lists from links bellow.
* Merged files already to download (from [config.yml](https://github.com/m0zgen/cactusd/blob/dev/config.yml)).

![today](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_date.svg)

* ![bl total](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_total_bl.svg) [bl.txt](https://raw.githubusercontent.com/m0zgen/cactusd/data/bl.txt) - Common merged block lists
* ![bl plain total](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_total_bl_plain.svg) [bl_plain.txt](https://raw.githubusercontent.com/m0zgen/cactusd/data/bl_plain.txt) - Merged lists from 
[BLD Agregator](https://github.com/m0zgen/bld-agregator) and [dns-hole](https://github.com/m0zgen/dns-hole) regex files
* ![wo total](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_total_wl.svg) [wl.txt](https://raw.githubusercontent.com/m0zgen/cactusd/data/wl.txt) - Common whitelisted domains from [dns-hole](https://github.com/m0zgen/dns-hole)
* ![wl plain total](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_total_wl_plain.svg) [wl_plain.txt](https://raw.githubusercontent.com/m0zgen/cactusd/data/wl_plain.txt) - Regex lists from [dns-hole](https://github.com/m0zgen/dns-hole)
* ![IP total](https://raw.githubusercontent.com/m0zgen/cactusd/data/badge_total_ip.svg) [ip_plain.txt](https://github.com/m0zgen/cactusd/raw/data/ip_plain.txt) - IP Block List (
C&C servers associated with Dridex, Emotet, TrickBot, QakBot, BazarLoader, and spammers/malicious IP addresses from Internet).

These lists updating every 12 hours.

If you are using `open-bld` or `blocky` DNS blocking servers, you can use these lists in `config.yml`, example:

```shell
...
blocking:
  blackLists:
    bld:
      - https://raw.githubusercontent.com/m0zgen/cactusd/data/bl.txt
      - https://raw.githubusercontent.com/m0zgen/cactusd/data/bl_plain.txt
  whiteLists:
    bld:
      - https://raw.githubusercontent.com/m0zgen/cactusd/data/wl.txt
      - https://raw.githubusercontent.com/m0zgen/cactusd/data/wl_plain.txt
...
```

You can combine blocking solution with `firewalld` and `ipset` with [ip2drop](https://github.com/m0zgen/ip2drop) 
configuration, example config for malicious IP relaxator: 

`conf.d/relax-ip.ini`
```shell
[DEFAULT]
IP_TIMEOUT = 604800
IP_THRESHOLD = -1
EXPORT_COMMAND = curl -s https://raw.githubusercontent.com/m0zgen/cactusd/data/ip_plain.txt
EXPORT_LOG = relax-ip.log
ENABLED = Yes
GROUP_NAME = relax-ip
EXPORT_TO_UPLOAD = No
DROP_DIRECTLY = Yes
```

## Self-Hosting Configuration

### Server Config
* `port` - Web severer port listening 
* `update_interval` - Heart beat in minutes (like as 30m)
* `download_dir` - lists download catalog
* `upload_dir` - catalog for remote file uploading
* `public_dir` - public web folder for downloaded, uploaded and merged files

## Lists Config

Block, White lists contains DNS names usually usage for DNS servers like as 
ad-guard, pi-hole, [open bld](https://lab.sys-adm.in) and etc.)

IP list - merging and aggregating IP lists from different sources (like as [bld-agregator](https://github.com/m0zgen/bld-agregator), [bld-server](https://github.com/m0zgen/bld-server))

Conditionally the `lists` are divided into several categories:
* `bl`, `wl` - blocking/white lists, hosts list with comments which 
need to clean and merge in solid file fo reducing size, remote server requests
* `bl_plain`, `wl_plain`, `ip_plain` - lists juts merging and clean empty spaces and lines and hashtags.

If you not need some list category, like as `wl_plain` or `ip_plain` just set `none` parameter:

```yaml
...
  wl_plain:
    - none
  ip_plain:
    - none
```

Every category will merge and publish in finally in `publish/files` catalog as solid files:
* `public/files/bl.txt`
* `public/files/wl.txt`
* `public/files/bl_plain.txt` - usually regex-based allowing lists for DNS
* `public/files/wl_plain.txt` - usually regex-based allowing/exception lists for DNS
* `public/files/ip_plain.txt` - blocking IP addresses (like example for [ip2drop](https://github.com/m0zgen/ip2drop) scripts or just for `ipset` blocking) 
* `public/files/dropped_ip.txt` - from remote [ip2drop](https://github.com/m0zgen/ip2drop) servers, oe any another script or routines

## Run Cactusd

From terminal:

```shell
./cactusd -config config.yml
```

From `systemd`:

```shell
...
#
ExecStart=/path/to/cactusd --config config-prod.yml
...
```

## Additional Info

* all functionality moved from deprecated [BLD-Server](https://github.com/m0zgen/bld-server)
* as "client" for `cactusd` server you can use [ip2drop](https://github.com/m0zgen/ip2drop) 
* [DNS-Hole](https://github.com/m0zgen/dns-hole) list collections 
* one more aggregator [BLD-Agregator](https://github.com/m0zgen/bld-agregator)
