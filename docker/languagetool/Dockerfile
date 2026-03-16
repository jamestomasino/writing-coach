FROM eclipse-temurin:21-jre

ARG LANGUAGETOOL_VERSION=6.6

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl unzip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/languagetool

RUN curl -fsSL "https://languagetool.org/download/LanguageTool-${LANGUAGETOOL_VERSION}.zip" -o /tmp/languagetool.zip \
    && unzip /tmp/languagetool.zip -d /opt \
    && rm /tmp/languagetool.zip \
    && mv /opt/LanguageTool-${LANGUAGETOOL_VERSION}/* /opt/languagetool \
    && rmdir /opt/LanguageTool-${LANGUAGETOOL_VERSION}

EXPOSE 8010

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=5 \
  CMD curl -fsS -d "language=en-US" -d "text=Health check." http://127.0.0.1:8010/v2/check >/dev/null || exit 1

CMD ["java", "-cp", "languagetool-server.jar:libs/*", "org.languagetool.server.HTTPServer", "--port", "8010", "--allow-origin"]
