#!/usr/bin/env python3

from pyinfra import host
from pyinfra.operations import apt, server, files, systemd

SUDO = True

if not host.data.LL_DOMAIN:
    raise RuntimeError("Define LL_DOMAIN")

apt.packages(
    name="Ensure all relevant apt packages",
    packages=["nginx", "fail2ban", "certbot", "python3-certbot-nginx"])

server.user("_ll",
            shell="/usr/sbin/nologin",
            group="_ll",
            system=True,
            ensure_home=False)

files.directory(name="Ensure data directory is present",
                path="/var/lib/ll",
                user="_ll",
                group="_ll",
                mode=750,
                recursive=True)

files.put(name="Upload daemon binary",
          src="../ll",
          dest="/usr/local/bin/ll",
          user="root",
          group="root")

files.put(name="Upload configuration file",
          src="ll.conf",
          dest="/etc/ll.conf",
          user="root",
          group="root")

files.put(name="Upload systemd unit file",
          src="ll.service",
          dest="/etc/systemd/system/sp.service",
          user="root",
          group="root")

systemd.daemon_reload()

systemd.service(name="Ensure ll service is enabled and running",
                service="ll.service",
                enabled=True,
                running=True,
                restarted=True)

files.put(name="NGINX configuration",
          src="yourdomain",
          dest=f"/etc/nginx/sites-available/{host.data.LL_DOMAIN}",
          user="root",
          group="root")
