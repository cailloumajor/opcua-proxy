FROM denoland/deno:distroless-2.6.9

WORKDIR /app

COPY deno.json *.ts ./

HEALTHCHECK --interval=5s --timeout=5s \
    CMD [ "deno", "run", "--allow-net", "/app/healthcheck.ts", "3000" ]

EXPOSE 3000
CMD ["serve", "--port", "3000", "/app/config-api-server.ts"]
