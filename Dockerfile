FROM postgres:latest

ENV POSTGRES_USER=${DB_USER}
ENV POSTGRES_PASSWORD=${DB_PASSWORD}
ENV POSTGRES_DB=${DB_NAME}

EXPOSE 5432

HEALTHCHECK --interval=5s --timeout=5s --retries=3 \
  CMD pg_isready -U "${DB_USER}" -d "${DB_NAME}"
