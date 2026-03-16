#!/bin/sh
set -eu

mode="${1:-serve}"

kratos_public_base_url="${KRATOS_PUBLIC_BASE_URL:-http://127.0.0.1:4433/}"
kratos_admin_base_url="${KRATOS_ADMIN_BASE_URL:-http://kratos:4434/}"
kratos_ui_public_url="${KRATOS_UI_PUBLIC_URL:-http://127.0.0.1:4455}"
coach_public_url="${COACH_PUBLIC_URL:-${kratos_ui_public_url}}"
kratos_allowed_return_urls="${KRATOS_ALLOWED_RETURN_URLS:-${kratos_ui_public_url},${coach_public_url}}"
kratos_smtp_connection_uri="${KRATOS_SMTP_CONNECTION_URI:-smtps://test:test@mailslurper:1025/?skip_ssl_verify=true}"

IFS=','
set -- $kratos_allowed_return_urls
return_url_1="${1:-$kratos_ui_public_url}"
return_url_2="${2:-$coach_public_url}"
return_url_3="${3:-$kratos_ui_public_url/login}"

sed \
  -e "s|__KRATOS_PUBLIC_BASE_URL__|$kratos_public_base_url|g" \
  -e "s|__KRATOS_ADMIN_BASE_URL__|$kratos_admin_base_url|g" \
  -e "s|__KRATOS_UI_PUBLIC_URL__|$kratos_ui_public_url|g" \
  -e "s|__COACH_PUBLIC_URL__|$coach_public_url|g" \
  -e "s|__KRATOS_ALLOWED_RETURN_URL_1__|$return_url_1|g" \
  -e "s|__KRATOS_ALLOWED_RETURN_URL_2__|$return_url_2|g" \
  -e "s|__KRATOS_ALLOWED_RETURN_URL_3__|$return_url_3|g" \
  -e "s|__KRATOS_SMTP_CONNECTION_URI__|$kratos_smtp_connection_uri|g" \
  /etc/config/kratos/kratos.yml.tmpl > /tmp/kratos.yml

if [ "$mode" = "migrate" ]; then
  exec kratos -c /tmp/kratos.yml migrate sql -e --yes
fi

exec kratos serve -c /tmp/kratos.yml --watch-courier
