FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY docker/nlp-analyzer/requirements.txt /app/requirements.txt

RUN pip install --no-cache-dir -r /app/requirements.txt

COPY docker/nlp-analyzer/app.py /app/app.py

EXPOSE 8020

CMD ["python", "/app/app.py"]
