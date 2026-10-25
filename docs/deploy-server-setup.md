# Server setup for deployment

One-time steps for a fresh Ubuntu 24.04 server (2 GB RAM, 30 GB disk is enough). After this, every push to
`main` deploys itself through `.github/workflows/deploy.yml`.

Replace `<user>` with your SSH user and `agropal.example.com` with your domain.

## 1. Basics

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install -y git ufw fail2ban

sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Use a normal user, not root, and log in with an SSH key. In `/etc/ssh/sshd_config` set
`PasswordAuthentication no` and `PermitRootLogin no`, then `sudo systemctl restart ssh`.
Keep a second SSH session open while you change this, so a mistake does not lock you out.

## 2. Docker

Install Docker Engine and the compose plugin from https://docs.docker.com/engine/install/ubuntu/, then:

```sh
sudo usermod -aG docker <user>     # log out and back in afterwards
docker compose version
```

## 3. DNS

Point an A record for your domain at the server's IP. Caddy cannot get a certificate until it resolves.

## 4. Let the server read the repo

```sh
ssh-keygen -t ed25519 -f ~/.ssh/agropal_deploy -N ""
cat ~/.ssh/agropal_deploy.pub
```

On GitHub add that public key at repo, Settings, Deploy keys (read-only). Then tell git to use it:

```sh
cat >> ~/.ssh/config <<'CONF'
Host github.com
  IdentityFile ~/.ssh/agropal_deploy
  IdentitiesOnly yes
CONF

git clone git@github.com:abyanhmd/agropal.git ~/agropal
```

## 5. Production settings

```sh
cd ~/agropal
cp .env.prod.example .env
nano .env
```

Set `DOMAIN`, the database and RabbitMQ passwords, the SMTP provider, and two secrets of at least 32
characters each (`openssl rand -hex 32`). The `.env` stays on the server: it is gitignored, so deploys never
overwrite it.

## 6. First deploy by hand

Do this once to prove the stack works before automating it.

```sh
cd ~/agropal
export COMPOSE_PARALLEL_LIMIT=1
docker compose -f docker-compose.prod.yaml build
docker compose -f docker-compose.prod.yaml up -d
docker compose -f docker-compose.prod.yaml ps
docker compose -f docker-compose.prod.yaml logs -f api
```

Then open `https://<your domain>`.

## 7. Connect GitHub Actions

Make a key used only for deployments:

```sh
# on your own computer
ssh-keygen -t ed25519 -f agropal_actions -N ""
```

- Append `agropal_actions.pub` to `~/.ssh/authorized_keys` on the server.
- In the repo, Settings, Secrets and variables, Actions, add:
  - `SSH_HOST`: the server IP or domain
  - `SSH_USER`: the SSH user
  - `SSH_KEY`: the full contents of the private file `agropal_actions`
- Delete the private key file from your computer afterwards.

Push to `main`, or run the **Deploy** workflow by hand from the Actions tab.

## Low memory (2 GB)

If a build ends with `Killed` or exit code 137 it ran out of memory. Add swap on the server:

```sh
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
free -h
```

If the Next.js build still gets killed, add `NODE_OPTIONS=--max-old-space-size=1024` to the `ENV` line in the
build stage of `frontend/Dockerfile`.

## Backups

Nothing here is backed up by default. A simple nightly dump:

```sh
cd ~/agropal
docker compose -f docker-compose.prod.yaml exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' | gzip > ~/backup-$(date +%F).sql.gz
```

Put it in `crontab -e` and copy the files off the server now and then.

## Useful commands

```sh
docker compose -f docker-compose.prod.yaml ps              # what is running
docker compose -f docker-compose.prod.yaml logs -f api     # follow a service's logs
docker system df                                           # how much disk Docker uses
```
