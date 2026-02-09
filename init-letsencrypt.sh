#!/bin/bash

# Let's Encrypt certificate initialization script

if [ -z "$1" ]; then
  echo "Usage: ./init-letsencrypt.sh <domain> <email>"
  echo "Example: ./init-letsencrypt.sh example.com admin@example.com"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Usage: ./init-letsencrypt.sh <domain> <email>"
  echo "Example: ./init-letsencrypt.sh example.com admin@example.com"
  exit 1
fi

DOMAIN=$1
EMAIL=$2

echo "### Initializing Let's Encrypt for domain: $DOMAIN"

# Create temporary nginx config for initial certificate request
echo "### Creating temporary nginx configuration..."

# Start nginx and certbot containers
echo "### Starting services..."
docker compose up -d nginx certbot

# Wait for nginx to be ready
echo "### Waiting for nginx to be ready..."
sleep 5

# Request certificate
echo "### Requesting Let's Encrypt certificate for $DOMAIN..."
docker compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email $EMAIL \
  --agree-tos \
  --no-eff-email \
  -d $DOMAIN

if [ $? -eq 0 ]; then
  echo ""
  echo "### Certificate obtained successfully!"
  echo "### Next steps:"
  echo "1. Update nginx/conf.d/default.conf:"
  echo "   - Uncomment the HTTPS server block"
  echo "   - Replace 'your-domain.com' with '$DOMAIN'"
  echo "2. Restart nginx: docker compose restart nginx"
  echo ""
  echo "### Certificate will auto-renew every 12 hours"
else
  echo ""
  echo "### Certificate request failed!"
  echo "### Make sure:"
  echo "1. Your domain '$DOMAIN' points to this server's public IP"
  echo "2. Ports 80 and 443 are accessible from the internet"
  echo "3. No firewall is blocking the ports"
fi
