# go-ssh-block-ip-bruteforce
Simple SSH IP blocker. Like Fail2ban but in Golang.

If a Redis server is available, then the application can distribute blocked IP addresses across servers and perform a massive ban.

For personal learning

WIP!!!

## Usign
Rename config.yml.example and load parameters.
You can use Redis.io to get access a free database.

## Feat:
- Parse some log file for failed access attempts.
- Save IP addresses in memory, counting attempts and last seen.
- Run local firewall command to deny IP addresses.
- Distributed ban IP using Redis.

## TODO:
- Unban IPs
- A lot of refactoring.
- ..etc
