# Let's Encrypt SSL Setup

## Prerequisites

1. A domain name pointing to your server's public IP
2. Ports 80 and 443 accessible from the internet
3. Docker and Docker Compose running

## Quick Start

### 1. Obtain Certificate

Run the initialization script with your domain and email:

```bash
./init-letsencrypt.sh your-domain.com your-email@example.com
```

Example:
```bash
./init-letsencrypt.sh example.com admin@example.com
```

### 2. Configure HTTPS

After successfully obtaining the certificate:

1. Edit `nginx/conf.d/default.conf`
2. Uncomment the HTTPS server block (lines starting with `# server {`)
3. Replace `your-domain.com` with your actual domain
4. Restart nginx:
   ```bash
   docker compose restart nginx
   ```

### 3. Test Your Setup

Visit `https://your-domain.com` to verify the SSL certificate is working.

## Certificate Auto-Renewal

The certbot container automatically checks for certificate renewal every 12 hours. No manual intervention needed!

## Manual Certificate Renewal

To manually renew certificates:

```bash
docker compose run --rm certbot renew
docker compose exec nginx nginx -s reload
```

## Troubleshooting

### Certificate Request Fails

Common issues:

1. **Domain not pointing to server**: Verify DNS with `dig your-domain.com`
2. **Firewall blocking**: Ensure ports 80/443 are open
3. **nginx not running**: Check with `docker compose ps nginx`

### View Certbot Logs

```bash
docker compose logs certbot
```

## Files and Directories

- `certbot_conf` volume: `/etc/letsencrypt` containing certificates
- `certbot_www` volume: `/var/www/certbot` for ACME challenges
- Certificates path: `/etc/letsencrypt/live/your-domain.com/`
  - `fullchain.pem` - Full certificate chain
  - `privkey.pem` - Private key

## Testing Locally (Self-Signed Certificate)

For development/testing without a real domain:

```bash
# Generate self-signed certificate
mkdir -p ./nginx/certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ./nginx/certs/selfsigned.key \
  -out ./nginx/certs/selfsigned.crt \
  -subj "/CN=localhost"

# Update nginx config to use:
# ssl_certificate /etc/nginx/certs/selfsigned.crt;
# ssl_certificate_key /etc/nginx/certs/selfsigned.key;
```

Then add `./nginx/certs:/etc/nginx/certs:ro` to nginx volumes in docker-compose.yml.
